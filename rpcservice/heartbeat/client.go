package heartbeat

import (
	"cbsignal/rpcservice"
	"github.com/lexkong/log"
	"net/rpc"
	"time"
)

const (
	HEARTBEAT_SERVICE = "HeartbeatService"
	PONG = ".Pong"
	PING_INTERVAL = 10
)

type Req struct {
	Addr string
}

type Client struct {
	*rpc.Client
	masterAddr string
	selfAddr string
	Peers      map[string]*rpcservice.Peer
}

func NewHeartbeatClient(masterAddr, selfAddr string) *Client {
	client := Client{
		masterAddr: masterAddr,
		selfAddr: selfAddr,
		Peers:  make(map[string]*rpcservice.Peer),
	}
	// 定时删除过期节点
	go func() {
		for {
			time.Sleep(CHECK_INTERVAL*time.Second)
			now := time.Now().Unix()
			//log.Infof("check peer ts")
			for addr, peer := range client.Peers {
				//log.Infof("now %d check peer ts %d", now, peer.Ts())
				if now - peer.Ts() > EXPIRE_TOMEOUT {
					// peer 过期
					log.Warnf("peer %s expired, delete", addr)
					delete(client.Peers, addr)
				}
			}
		}
	}()
	return &client
}

func (h *Client) DialHeartbeatService() {
	if h.masterAddr == "" {
		panic("masterAddr is nil")
	}
	for {
		c, err := rpc.Dial("tcp", h.masterAddr)
		if err != nil {
			log.Errorf(err, "DialHeartbeatService")
			time.Sleep(5*time.Second)
			continue
		}
		h.Client = c
		break
	}
}

func (h *Client) sendMsgPing(request Req, reply *Resp) error {
	return h.Client.Call(HEARTBEAT_SERVICE+PONG, request, reply)
}

func (h *Client) StartHeartbeat() {
	go func() {
		for {
			time.Sleep(PING_INTERVAL*time.Second)
			heartbeatReq := Req{
				Addr: h.selfAddr,
			}
			var heartbeatResp Resp
			if err := h.sendMsgPing(heartbeatReq, &heartbeatResp);err != nil {
				log.Errorf(err, "heartbeat")
				h.DialHeartbeatService()
			}
			log.Infof("heartbeatResp %s", heartbeatResp)
			for _, addr := range heartbeatResp.Nodes {
				p, ok := h.Peers[addr]
				if ok {
					p.UpdateTs()
				} else {
					peer := rpcservice.NewPeer(addr)
					if err := peer.DialPeers();err != nil {
						h.Peers[addr] = peer
					}
				}
			}
		}
	}()
}