package heartbeat

import (
	"cbsignal/client"
	"cbsignal/hub"
	"cbsignal/rpcservice"
	"github.com/lexkong/log"
	"net/rpc"
	"time"
)

const (
	CHECK_INTERVAL = 25
	EXPIRE_TOMEOUT = 21
)


type PongResp struct {
	Nodes []string
}

type Peer struct {
	PeerId          string              //唯一标识
	RpcNodeAddr string       // rpc节点id
}

type PeersResp struct {
	Peers []*Peer
}

type Service struct {
	Nodes map[string]*rpcservice.Node // master维护的node集合
}

func RegisterHeartbeatService() error {
	log.Infof("register rpc service %s", HEARTBEAT_SERVICE)
	s := new(Service)
	s.Nodes = make(map[string]*rpcservice.Node)
	// 定时删除过期节点
	go func() {
		for {
			time.Sleep(CHECK_INTERVAL*time.Second)
			now := time.Now().Unix()
			//log.Infof("check node ts")
			for addr, node := range s.Nodes {
				//log.Infof("now %d check node ts %d", now, node.Ts())
				if now - node.Ts() > EXPIRE_TOMEOUT {
					// node 过期
					log.Warnf("node %s expired, delete", addr)
					delete(s.Nodes, addr)
				}
			}
		}
	}()
	return rpc.RegisterName(HEARTBEAT_SERVICE, s)
}

func (h *Service) Pong(request PingReq, reply *PongResp) error {
	addr := request.Addr
	//log.Infof("receive ping from %s", addr)
	p, ok := h.Nodes[addr]
	if ok {
		p.UpdateTs()
	} else {
		h.Nodes[addr] = rpcservice.NewNode(addr)
	}

	for key, _ := range h.Nodes {
		if key != addr {
			reply.Nodes = append(reply.Nodes, key)
		}
	}
	return nil
}

func (h *Service)Peers(request GetPeersReq, reply *PeersResp) error {
	var peers []*Peer
	//hub.GetInstance().Clients.Range(func(key, value interface{}) bool {
	//	cli := value.(*client.Client)
	//	peer := Peer{
	//		PeerId:      cli.PeerId,
	//		RpcNodeAddr: cli.RpcNodeAddr,
	//	}
	//	peers = append(peers, &peer)
	//	return true
	//})
	for item := range hub.GetInstance().Clients.IterBuffered() {
		cli := item.Val.(*client.Client)
		peer := Peer{
			PeerId:      cli.PeerId,
			RpcNodeAddr: cli.RpcNodeAddr,
		}
		peers = append(peers, &peer)
	}
	log.Infof("send %d peers to %s", len(peers), request.Addr)
	reply.Peers = peers
	return nil
}
