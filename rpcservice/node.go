package rpcservice

import (
	"cbsignal/rpcservice/pool"
	"errors"
	"fmt"
	"github.com/lexkong/log"
	"net/rpc"
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
	POOL_MIN_CONNS = 5
	POOL_MAX_CONNS = 50
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
	p, err := pool.NewGenericPool(POOL_MIN_CONNS, POOL_MAX_CONNS, time.Minute*10, func() (*rpc.Client, error) {
		c, err := rpc.Dial("tcp", addr)
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
	client, err := s.connPool.Acquire()
	if err != nil {
		return err
	}
	err = s.sendInternal(BROADCAST_SERVICE+JOIN, request, reply, client)
	s.connPool.Release(client)
	return err
}

func (s *Node) SendMsgLeave(request JoinLeaveReq, reply *RpcResp) error {
	if !s.isAlive {
		return errors.New(fmt.Sprintf("node %s is not alive", s.addr))
	}
	log.Infof("SendMsgLeave to %s", s.addr)
	request.Addr = s.addr
	client, err := s.connPool.Acquire()
	if err != nil {
		return err
	}
	err = s.sendInternal(BROADCAST_SERVICE+LEAVE, request, reply, client)
	s.connPool.Release(client)
	return err
}

func (s *Node) SendMsgSignal(request SignalReq, reply *RpcResp) error {
	if !s.isAlive {
		return errors.New(fmt.Sprintf("node %s is not alive", s.addr))
	}
	//log.Infof("SendMsgSignal to %s", s.addr)
	client, err := s.connPool.Acquire()
	if err != nil {
		return err
	}
	err = s.sendInternal(SIGNAL_SERVICE+SIGNAL, request, reply, client)
	s.connPool.Release(client)
	return err
}

func (s *Node) SendMsgPing(request Ping, reply *Pong) error {
	if !s.isAlive {
		return errors.New(fmt.Sprintf("node %s is not alive", s.addr))
	}
	//log.Infof("SendMsgPing to %s", s.addr)
	client, err := s.connPool.Acquire()

	if err != nil {
		return err
	}
	err = s.sendInternal(BROADCAST_SERVICE+PONG, request, reply, client)
	s.connPool.Release(client)
	return err
}

func (s *Node) sendInternal(method string, args interface{}, reply interface{}, client *rpc.Client ) error {
	//start := time.Now()
	//done := make(chan error, 1)

	done := make(chan *rpc.Call, 1)

	//log.Warnf("GenericPool now conn %d idle %d", s.connPool.NumTotalConn(), s.connPool.NumIdleConn())

	//go func() {
	//	//log.Warnf("client.Call %s", method)
	//	err := client.Call(method, args, reply)
	//	s.connPool.Release(closer)
	//	done <- err
	//}()

	client.Go(method, args, reply, done)
	
	//s.Client.Go(method, args, reply, done)
	//return call.Error

	select {
		case <-time.After(READ_TIMEOUT):
			//log.Warnf("rpc call timeout %s", method)
			//s.Client.Close()
			return fmt.Errorf("rpc call timeout %s", method)
		case call := <-done:
			//elapsed := time.Since(start)
			//log.Warnf("6666 %d %d", elapsed.Nanoseconds(), PRINT_WARN_LIMIT_NANO)
			//if elapsed.Nanoseconds() >= PRINT_WARN_LIMIT_NANO {
			//	log.Warnf("rpc send %s cost %v", method, elapsed)
			//}

			//if err.Error != nil {
			//	//rpcClient.Close()
			//	return err.Error
			//}
			if err := call.Error; err != nil {
				//rpcClient.Close()
				return err
			}
	}

	return nil
}

func (s *Node) StartHeartbeat() {
	go func() {
		for {
			log.Warnf("GenericPool now conn %d idle %d", s.connPool.NumTotalConn(), s.connPool.NumIdleConn())
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
