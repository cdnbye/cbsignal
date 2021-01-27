package rpcservice

import (
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
	ATTENTS_INTERVAL  = 2     // second
	PING_INTERVAL     = 5
)

type JoinLeaveReq struct {
	PeerId string                // 节点id
	Addr   string
}

type RpcResp struct {
	Success bool
	Reason string
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
	*rpc.Client
	addr string         // ip:port
	ts int64
	isAlive bool                 // 是否存活
}

func NewNode(addr string) *Node {
	node := Node{
		addr: addr,
		ts: time.Now().Unix(),
	}
	return &node
}

func (s *Node) DialNode() error {
	attemts := 0
	for {
		c, err := rpc.Dial("tcp", s.addr)
		if err != nil {
			if attemts >= DIAL_MAX_ATTENTS {
				return err
			}
			attemts ++
			log.Errorf(err, "DialNode")
			time.Sleep(ATTENTS_INTERVAL*time.Second)
			continue
		}
		s.Client = c
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
	return s.Client.Call(BROADCAST_SERVICE+JOIN, request, reply)
}

func (s *Node) SendMsgLeave(request JoinLeaveReq, reply *RpcResp) error {
	if !s.isAlive {
		return errors.New(fmt.Sprintf("node %s is not alive", s.addr))
	}
	log.Infof("SendMsgLeave to %s", s.addr)
	request.Addr = s.addr
	return s.Client.Call(BROADCAST_SERVICE+LEAVE, request, reply)
}

func (s *Node) SendMsgSignal(request SignalReq, reply *RpcResp) error {
	if !s.isAlive {
		return errors.New(fmt.Sprintf("node %s is not alive", s.addr))
	}
	//log.Infof("SendMsgSignal to %s", s.addr)
	return s.Client.Call(SIGNAL_SERVICE+SIGNAL, request, reply)
}

func (s *Node) SendMsgPing(request Ping, reply *Pong) error {
	if !s.isAlive {
		return errors.New(fmt.Sprintf("node %s is not alive", s.addr))
	}
	//log.Infof("SendMsgPing to %s", s.addr)
	return s.Client.Call(BROADCAST_SERVICE+PONG, request, reply)
}

func (s *Node) StartHeartbeat() {
	go func() {
		for {
			ping := Ping{}
			var pong Pong
			if err := s.SendMsgPing(ping, &pong);err != nil {
				log.Errorf(err, "node heartbeat")
				s.Lock()
				s.isAlive = false
				s.Unlock()
				if err := s.DialNode();err != nil {
					log.Errorf(err, "dial node")
					break
				}
			}
			time.Sleep(PING_INTERVAL*time.Second)
		}
	}()
}

func (s *Node) IsAlive() bool {
	return s.isAlive
}

