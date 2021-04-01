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
	PRINT_WARN_LIMIT_NANO = 100 * time.Millisecond
	POOL_MIN_CONNS = 5
	POOL_MAX_CONNS = 32
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

type Node struct {
	sync.Mutex
	addr             string // ip:port
	ts               int64
	isAlive          bool // 是否存活
	connPool         pool.Pool
	Released         bool
}

//func NewNode(addr string) *Node {
//	node := Node{
//		addr: addr,
//		ts:   time.Now().Unix(),
//	}
//	// 创建连接池
//	p, err := pool.NewGenericPool(POOL_MIN_CONNS, POOL_MAX_CONNS, time.Minute*10, func() (*rpc.Client, error) {
//		c, err := rpc.Dial("tcp", addr)
//		if err != nil {
//
//			return nil, err
//		}
//		return c, nil
//	})
//	if err != nil {
//		panic(err)
//	}
//	node.connPool = p
//	return &node
//}

func NewNode(addr string) *Node {
	node := Node{
		addr: addr,
		ts:   time.Now().Unix(),
	}

	//factory 创建连接的方法
	factory := func() (*rpc.Client, error) {
		c, err := rpc.Dial("tcp", addr)
		if err != nil {

			return nil, err
		}
		return c, nil
	}

	//close 关闭连接的方法
	closer := func(v *rpc.Client) error { return v.Close() }

	poolConfig := &pool.Config{
		InitialCap: POOL_MIN_CONNS,         //资源池初始连接数
		MaxIdle:   POOL_MAX_CONNS,                 //最大空闲连接数
		MaxCap:     POOL_MAX_CONNS,//最大并发连接数
		Factory:    factory,
		Close:      closer,
		//连接最大空闲时间，超过该时间的连接 将会关闭，可避免空闲时连接EOF，自动失效的问题
		IdleTimeout: 60 * time.Second,
	}
	p, err := pool.NewChannelPool(poolConfig)

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

	return s.sendMsg(BROADCAST_SERVICE+JOIN, request, reply)
}

func (s *Node) SendMsgLeave(request JoinLeaveReq, reply *RpcResp) error {
	if !s.isAlive {
		return errors.New(fmt.Sprintf("node %s is not alive", s.addr))
	}
	log.Infof("SendMsgLeave to %s", s.addr)
	request.Addr = s.addr
	return s.sendMsg(BROADCAST_SERVICE+LEAVE, request, reply)
}

func (s *Node) SendMsgSignal(request SignalReq, reply *RpcResp) error {
	if !s.isAlive {
		return errors.New(fmt.Sprintf("node %s is not alive", s.addr))
	}
	//log.Infof("SendMsgSignal to %s", s.addr)

	return s.sendMsg(SIGNAL_SERVICE+SIGNAL, request, reply)

}

func (s *Node) SendMsgPing(request Ping, reply *Pong) error {
	//log.Infof("SendMsgPing to %s", s.addr)
	return s.sendMsg(BROADCAST_SERVICE+PONG, request, reply)
}

func (s *Node) sendMsg(method string, request interface{}, reply interface{}) error {
	client, err := s.connPool.Acquire()

	if err != nil {
		return err
	}
	err = s.sendInternal(method, request, reply, client)
	if err != nil {
		s.connPool.Close(client)
	} else {
		s.connPool.Release(client)
	}
	return err
}

func (s *Node) sendInternal(method string, request interface{}, reply interface{}, client *rpc.Client) error {
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

	client.Go(method, request, reply, done)
	//s.connPool.Release(client)
	//s.Client.Go(method, args, reply, done)
	//return call.Error

	select {
		case <-time.After(READ_TIMEOUT):
			//log.Warnf("rpc call timeout %s", method)
			//s.Client.Close()
			return fmt.Errorf("rpc call timeout %s", method)
		case call := <-done:
			//.Add(timeout).Before(time.Now())
			//elapsed := time.Since(start)
			//log.Warnf("6666 %d %d", elapsed.Nanoseconds(), PRINT_WARN_LIMIT_NANO)
			//if start.Add(PRINT_WARN_LIMIT_NANO).Before(time.Now()) {
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
			if s.Released {
				//log.Warnf("%s s.Released", s.addr)
				break
			}
			log.Warnf("ConnPool %s conn %d idle %d", s.addr, s.connPool.NumTotalConn(), s.connPool.NumIdleConn())
			ping := Ping{}
			var pong Pong
			if err := s.SendMsgPing(ping, &pong); err != nil {
				log.Errorf(err, "node heartbeat")
				s.Lock()
				s.isAlive = false
				s.Unlock()
			} else {
				s.Lock()
				s.isAlive = true
				s.Unlock()
			}
			time.Sleep(PING_INTERVAL * time.Second)
		}
	}()
}

func (s *Node) IsAlive() bool {
	return s.isAlive
}
