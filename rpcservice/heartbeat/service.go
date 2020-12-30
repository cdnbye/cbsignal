package heartbeat

import (
	"cbsignal/rpcservice"
	"github.com/lexkong/log"
	"net/rpc"
	"time"
)

const (
	CHECK_INTERVAL = 30
	EXPIRE_TOMEOUT = 21
)


type Resp struct {
	Nodes []string
}

type HeartbeatService struct {
	Peers map[string]*rpcservice.Peer
}

func RegisterHeartbeatService() error {
	log.Infof("register rpcservice service %s", HEARTBEAT_SERVICE)
	s := new(HeartbeatService)
	s.Peers = make(map[string]*rpcservice.Peer)
	// 定时删除过期节点
	go func() {
		for {
			time.Sleep(CHECK_INTERVAL*time.Second)
			now := time.Now().Unix()
			//log.Infof("check peer ts")
			for addr, peer := range s.Peers {
				//log.Infof("now %d check peer ts %d", now, peer.Ts())
				if now - peer.Ts() > EXPIRE_TOMEOUT {
					// peer 过期
					log.Warnf("peer %s expired, delete", addr)
					delete(s.Peers, addr)
				}
			}
		}
	}()
	return rpc.RegisterName(HEARTBEAT_SERVICE, s)
}

func (h *HeartbeatService) Pong(request Req, reply *Resp) error {
	addr := request.Addr
	log.Infof("HeartbeatService receive %s", addr)
	p, ok := h.Peers[addr]
	if ok {
		p.UpdateTs()
	} else {
		h.Peers[addr] = rpcservice.NewPeer(addr)
	}

	for key, _ := range h.Peers {
		if key != addr {
			reply.Nodes = append(reply.Nodes, key)
		}
	}

	return nil
}
