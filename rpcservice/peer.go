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
	DIAL_MAX_ATTENTS = 3
	ATTENTS_INTERVAL = 2     // second
)

type JoinLeaveReq struct {
	Id string           // 节点id
	Addr string
}

type JoinLeaveResp struct {
	Success bool
}

type Peer struct {
	*rpc.Client
	addr string         // ip:port
	ts int64
}

func NewPeer(addr string) *Peer {
	peer := Peer{
		addr:addr,
		ts: time.Now().Unix(),
	}
	return &peer
}

func (s *Peer) DialPeers() error {
	attemts := 0
	for {
		c, err := rpc.Dial("tcp", s.addr)
		if err != nil {
			if attemts > DIAL_MAX_ATTENTS {
				return err
			}
			attemts ++
			log.Errorf(err, "DialHeartbeatService")
			time.Sleep(ATTENTS_INTERVAL*time.Second)
			continue
		}
		s.Client = c
		break
	}
	return nil
}

func (s *Peer) UpdateTs() {
	s.ts = time.Now().Unix()
}

func (s *Peer) IsMaster() bool {
	return false
}

func (s *Peer) Addr() string {
	return s.addr
}

func (s *Peer) Ts() int64 {
	return s.ts
}

func (s *Peer) SendMsgJoin(request JoinLeaveReq, reply *JoinLeaveResp) error {
	request.Addr = s.addr
	return s.Client.Call(BROADCAST_SERVICE+JOIN, request, reply)
}

func (s *Peer) SendMsgLeave(request JoinLeaveReq, reply *JoinLeaveResp) error {
	request.Addr = s.addr
	return s.Client.Call(BROADCAST_SERVICE+LEAVE, request, reply)
}

