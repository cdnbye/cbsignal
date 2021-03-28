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
	cli := Client{
		masterAddr: masterAddr,
		selfAddr: selfAddr,
		nodeHub: rpcservice.NewNodeHub(),
	}
	return &cli
}

func (h *Client) NodeHub() *rpcservice.NodeHub {
	return h.nodeHub
}

// 连接master并向master获取节点
func (h *Client) DialHeartbeatService() {
	if h.masterAddr == "" {
		panic("masterAddr is nil")
	}
	for {
		c, err := rpc.Dial("tcp", h.masterAddr)
		if err != nil {
			log.Errorf(err, "DialHeartbeatService")
			// 与master失去联系，清空所有peers
			//hub.ClearAll()
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
		if peer.RpcNodeAddr == h.selfAddr {
			// 本节点的peer忽略
			continue
		}
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
				// master失去联系
				h.IsMasterAlive = false
				// 将master节点对应的peers删除
				go deletePeersInNode(h.masterAddr)
				h.DialHeartbeatService()
			}
			log.Debugf("heartbeatResp %s", heartbeatResp)
			// 删除死节点
			for addr, node := range h.nodeHub.GetAll() {
				if !node.IsAlive() {
					h.nodeHub.Delete(addr)
					// 将节点对应的peers删除
					go deletePeersInNode(addr)
				}
			}
			for _, addr := range heartbeatResp.Nodes {
				_, ok := h.nodeHub.Get(addr)
				if !ok {
					log.Infof("New Node %s", addr)
					node := rpcservice.NewNode(addr)
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



func deletePeersInNode(addr string)  {
	//hub.GetInstance().Clients.Range(func(peerId, peer interface{}) bool {
	//	cli := peer.(*client.Client)
	//	if cli.RpcNodeAddr == addr {
	//		log.Infof("delete cli %s in deleted node %s", cli.PeerId, addr)
	//		hub.DoUnregister(cli.PeerId)
	//	}
	//	return true
	//})

	for item := range hub.GetInstance().Clients.IterBuffered() {
		val := item.Val
		cli := val.(*client.Client)
		if cli.RpcNodeAddr == addr {
			log.Infof("delete cli %s in deleted node %s", cli.PeerId, addr)
			hub.DoUnregister(cli.PeerId)
		}
	}
}