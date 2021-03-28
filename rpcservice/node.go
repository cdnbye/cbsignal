package rpcservice

import (
	"cbsignal/rpcservice/pool"
	"errors"
	"fmt"
	"github.com/lexkong/log"
	"io"
	"net/rpc"
	"strings"
	"sync"
	"time"
)

const (
	BROADCAST_SERVICE = "BroadcastService"
	JOIN              = ".Join"
	LEAVE             = ".Leave"
	PONG              = ".Pong"
	SIGNAL_SERVICE    = "SignalService"
	SIGNAL            = ".Signal"
	DIAL_MAX_ATTENTS  = 2
	ATTENTS_INTERVAL  = 2 // second
	PING_INTERVAL     = 5
	DIAL_TIMEOUT      = 3    // second
	READ_TIMEOUT      = 1500 * time.Millisecond
	PRINT_WARN_LIMIT_NANO = 100 * 1000000
)



type JoinLeaveReq struct {
	PeerId string // 节点id
	Addr   string
}

type RpcResp struct {
	Success bool
	Reason  string
}

type SignalReq struct {
	ToPeerId string
	Data     []byte
}

type Ping struct {
}

type Pong struct {
}

//type Conn struct {
//	*rpc.Client
//}

//func (c *Conn)close() error {
//	return c.Close()
//}

type Node struct {
	sync.Mutex
	msgChannel       *rpc.Client
	joinLeaveChannel *rpc.Client
	addr             string // ip:port
	ts               int64
	isAlive          bool // 是否存活
	connPool         *pool.GenericPool
}

func NewNode(addr string) *Node {
	node := Node{
		addr: addr,
		ts:   time.Now().Unix(),
	}
	// 创建连接池
	p, err := pool.NewGenericPool(1, 1, time.Minute*10, func() (io.Closer, error) {
		c, err := rpc.Dial("tcp", addr)
		log.Warnf("NewGenericPool rpc.Dial %s", addr)
		if err != nil {

			return nil, err
		}
		return c, nil
	})
	if err != nil {
		panic(err)
	}
	node.connPool = p
	return &node
}

func (s *Node)DialNode() error {
	s.Lock()
	s.isAlive = true
	s.Unlock()
	return nil
}

func (s *Node) dialMsgChannel() error {
	return s.dialWithChannel(s.msgChannel)
}

func (s *Node) dialJoinLeaveChannel() error {
	//return s.dialWithChannel(s.joinLeaveChannel)
	attemts := 0
	s.Lock()
	s.isAlive = false
	s.Unlock()
	joinLeaveAddr := strings.Split(s.addr, ":")[0]+":12000"
	for {
		c, err := rpc.Dial("tcp", joinLeaveAddr)
		if err != nil {
			if attemts >= DIAL_MAX_ATTENTS {
				return err
			}
			attemts++
			log.Errorf(err, "dialWithChannel")
			time.Sleep(ATTENTS_INTERVAL * time.Second)
			continue
		}
		s.joinLeaveChannel = c
		break
	}
	log.Warnf("Dial node %s succeed", joinLeaveAddr)
	s.Lock()
	s.isAlive = true
	s.Unlock()
	return nil
}

func (s *Node)dialWithChannel(client *rpc.Client) error {
	attemts := 0
	s.Lock()
	s.isAlive = false
	s.Unlock()
	for {
		c, err := rpc.Dial("tcp", s.addr)
		if err != nil {
			if attemts >= DIAL_MAX_ATTENTS {
				return err
			}
			attemts++
			log.Errorf(err, "dialWithChannel")
			time.Sleep(ATTENTS_INTERVAL * time.Second)
			continue
		}
		client = c
		break
	}
	log.Warnf("Dial node %s succeed", s.addr)
	s.Lock()
	s.isAlive = true
	s.Unlock()
	return nil
}

func (s *Node) UpdateTs() {
	s.ts = time.Now().Unix()
}

func (s *Node) IsMaster() bool {
	return false
}

func (s *Node) Addr() string {
	return s.addr
}

func (s *Node) Ts() int64 {
	return s.ts
}

func (s *Node) SendMsgJoin(request JoinLeaveReq, reply *RpcResp) error {
	if !s.isAlive {
		return errors.New(fmt.Sprintf("node %s is not alive", s.addr))
	}
	log.Infof("SendMsgJoin to %s", s.addr)
	return s.sendInternal(BROADCAST_SERVICE+JOIN, request, reply)
}

func (s *Node) SendMsgLeave(request JoinLeaveReq, reply *RpcResp) error {
	if !s.isAlive {
		return errors.New(fmt.Sprintf("node %s is not alive", s.addr))
	}
	log.Infof("SendMsgLeave to %s", s.addr)
	request.Addr = s.addr
	return s.sendInternal(BROADCAST_SERVICE+LEAVE, request, reply)
}

func (s *Node) SendMsgSignal(request SignalReq, reply *RpcResp) error {
	if !s.isAlive {
		return errors.New(fmt.Sprintf("node %s is not alive", s.addr))
	}
	//log.Infof("SendMsgSignal to %s", s.addr)
	return s.sendInternal(SIGNAL_SERVICE+SIGNAL, request, reply)
}

func (s *Node) SendMsgPing(request Ping, reply *Pong) error {
	if !s.isAlive {
		return errors.New(fmt.Sprintf("node %s is not alive", s.addr))
	}
	log.Infof("SendMsgPing to %s", s.addr)
	return s.sendInternal(BROADCAST_SERVICE+PONG, request, reply)
}

func (s *Node) sendInternal(method string, args interface{}, reply interface{}) error {
	//start := time.Now()
	done := make(chan error, 1)

	//done := make(chan *rpc.Call, 1)

	closer, err := s.connPool.Acquire()
	if err != nil {
		return err
	}
	client := closer.(*rpc.Client)
	go func() {
		log.Warnf("client.Call %s", method)
		err := client.Call(method, args, reply)
		done <- err
	}()

	//client.Go(method, args, reply, done)
	
	//s.Client.Go(method, args, reply, done)
	//return call.Error

	select {
		case <-time.After(READ_TIMEOUT):
			//log.Warnf("rpc call timeout %s", method)
			//s.Client.Close()
			s.connPool.Release(closer)
			return fmt.Errorf("rpc call timeout %s", method)
		case err := <-done:
			//elapsed := time.Since(start)
			//log.Warnf("6666 %d %d", elapsed.Nanoseconds(), PRINT_WARN_LIMIT_NANO)
			//if elapsed.Nanoseconds() >= PRINT_WARN_LIMIT_NANO {
			//	log.Warnf("rpc send %s cost %v", method, elapsed)
			//}

			//if err.Error != nil {
			//	//rpcClient.Close()
			//	return err.Error
			//}
			s.connPool.Release(closer)
			if err != nil {
				//rpcClient.Close()
				return err
			}
	}

	return nil
}

func (s *Node) StartHeartbeat() {
	go func() {
		for {
			ping := Ping{}
			var pong Pong
			if err := s.SendMsgPing(ping, &pong); err != nil {
				log.Errorf(err, "node heartbeat")
				s.Lock()
				s.isAlive = false
				s.Unlock()
				if err := s.DialNode(); err != nil {
					log.Errorf(err, "dial node")
					break
				}
			}
			time.Sleep(PING_INTERVAL * time.Second)
		}
	}()
}

func (s *Node) IsAlive() bool {
	return s.isAlive
}
