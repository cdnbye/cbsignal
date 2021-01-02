package heartbeat

import (
	"cbsignal/hub"
	"cbsignal/rpcservice"
	"github.com/lexkong/log"
	"net/rpc"
	"time"
)

const (
	HEARTBEAT_SERVICE = "HeartbeatService"
	PONG = ".Pong"
	PEERS = ".Peers"
	PING_INTERVAL = 10
)

type PingReq struct {
	Addr string
}

type GetPeersReq struct {
	Addr string
}

type Client struct {
	*rpc.Client
	masterAddr    string
	selfAddr      string
	nodeHub       *rpcservice.NodeHub
	IsMasterAlive bool
}

func NewHeartbeatClient(masterAddr, selfAddr string) *Client {
	if masterAddr == selfAddr {
		log.Warnf("This is master node")
	} else {
		log.Warnf("This is slave node")
	}
	client := Client{
		masterAddr: masterAddr,
		selfAddr: selfAddr,
		nodeHub: rpcservice.NewNodeHub(),
	}
	// 定时删除过期节点
	go func() {
		for {
			time.Sleep(CHECK_INTERVAL*time.Second)
			now := time.Now().Unix()
			//log.Infof("check peer ts")
			for addr, peer := range client.nodeHub.GetAll() {
				//log.Infof("now %d check peer ts %d", now, peer.Ts())
				if now - peer.Ts() > EXPIRE_TOMEOUT {
					// peer 过期
					log.Warnf("node %s expired, delete", addr)
					client.nodeHub.Delete(addr)
				}
			}
		}
	}()
	return &client
}

func (h *Client) NodeHub() *rpcservice.NodeHub {
	return h.nodeHub
}

func (h *Client) DialHeartbeatService() {
	if h.masterAddr == "" {
		panic("masterAddr is nil")
	}
	for {
		c, err := rpc.Dial("tcp", h.masterAddr)
		if err != nil {
			log.Errorf(err, "DialHeartbeatService")
			// 与master失去联系，清空所有peers
			hub.ClearAll()
			time.Sleep(5*time.Second)
			continue
		}
		h.Client = c
		h.IsMasterAlive = true
		break
	}
	// 获取master的所有peer节点
	req := GetPeersReq{
		Addr: h.selfAddr,
	}
	var resp PeersResp
	if err := h.sendMsgGetPeers(req, &resp);err != nil {
		panic(err)
	}
	log.Warnf("Got %d peers from master", len(resp.Peers))
	for _, peer := range resp.Peers {
		hub.DoRegisterRemoteClient(peer.PeerId, peer.RpcNodeAddr)
	}
}

func (h *Client) sendMsgPing(request PingReq, reply *PongResp) error {
	return h.Client.Call(HEARTBEAT_SERVICE+PONG, request, reply)
}

func (h *Client) sendMsgGetPeers(request GetPeersReq, reply *PeersResp) error {
	return h.Client.Call(HEARTBEAT_SERVICE+PEERS, request, reply)
}

func (h *Client) StartHeartbeat() {
	go func() {
		for {
			heartbeatReq := PingReq{
				Addr: h.selfAddr,
			}
			var heartbeatResp PongResp
			if err := h.sendMsgPing(heartbeatReq, &heartbeatResp);err != nil {
				log.Errorf(err, "heartbeat")
				// master失去联系，停止服务
				h.IsMasterAlive = false
				h.DialHeartbeatService()
			}
			log.Infof("heartbeatResp %s", heartbeatResp)
			// 删除死节点
			for addr, node := range h.nodeHub.GetAll() {
				if !node.IsAlive() {
					h.nodeHub.Delete(addr)
				}
			}
			for _, addr := range heartbeatResp.Nodes {
				p, ok := h.nodeHub.Get(addr)
				if ok {
					//log.Infof("update %s ts", p.Addr())
					p.UpdateTs()
				} else {
					log.Infof("NewPeer %s", addr)
					node := rpcservice.NewPeer(addr)
					if err := node.DialNode();err == nil {
						h.nodeHub.Add(addr, node)
						node.StartHeartbeat()
					}
				}
			}
			time.Sleep(PING_INTERVAL*time.Second)
		}
	}()
}