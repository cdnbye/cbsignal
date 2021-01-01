package rpcservice

import (
	"github.com/lexkong/log"
	"net/rpc"
	"time"
)

const (
	BROADCAST_SERVICE = "BroadcastService"
	JOIN = ".Join"
	LEAVE = ".Leave"
	SIGNAL_SERVICE = "SignalService"
	SIGNAL = ".Signal"
	DIAL_MAX_ATTENTS = 3
	ATTENTS_INTERVAL = 2     // second
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

type Node struct {
	*rpc.Client
	addr string         // ip:port
	ts int64
}

func NewPeer(addr string) *Node {
	peer := Node{
		addr: addr,
		ts: time.Now().Unix(),
	}
	return &peer
}

func (s *Node) DialPeers() error {
	attemts := 0
	for {
		c, err := rpc.Dial("tcp", s.addr)
		if err != nil {
			if attemts > DIAL_MAX_ATTENTS {
				return err
			}
			attemts ++
			log.Errorf(err, "DialPeer")
			time.Sleep(ATTENTS_INTERVAL*time.Second)
			continue
		}
		s.Client = c
		break
	}
	log.Warnf("Dial peer %s succeed", s.addr)
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
	log.Infof("SendMsgJoin to %s", s.addr)
	return s.Client.Call(BROADCAST_SERVICE+JOIN, request, reply)
}

func (s *Node) SendMsgLeave(request JoinLeaveReq, reply *RpcResp) error {
	log.Infof("SendMsgLeave to %s", s.addr)
	request.Addr = s.addr
	return s.Client.Call(BROADCAST_SERVICE+LEAVE, request, reply)
}

func (s *Node) SendMsgSignal(request SignalReq, reply *RpcResp) error {
	log.Infof("SendMsgSignal to %s", s.addr)
	return s.Client.Call(SIGNAL_SERVICE+SIGNAL, request, reply)
}

