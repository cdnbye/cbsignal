package broadcast

import (
	"cbsignal/rpcservice"
	"github.com/lexkong/log"
)

type Client struct {
	nodeHub *rpcservice.NodeHub
	selfAddr string
}

func NewBroadcastClient(nodeHub *rpcservice.NodeHub, addr string) *Client {
	client := Client{
		nodeHub: nodeHub,
		selfAddr: addr,
	}
	return &client
}

func (c *Client) BroadcastMsgJoin(id string)  {
	log.Infof("BroadcastMsgJoin %s", id)
	req := rpcservice.JoinLeaveReq{
		PeerId: id,
		Addr: c.selfAddr,
	}
	var resp rpcservice.RpcResp
	for _, peer := range c.nodeHub.GetAll() {
		if err := peer.SendMsgJoin(req, &resp); err != nil {
			log.Warnf("peer %s SendMsgJoin failed", peer.Addr())
		}
	}
}

func (c *Client) BroadcastMsgLeave(id string)  {
	req := rpcservice.JoinLeaveReq{
		PeerId: id,
	}
	var resp rpcservice.RpcResp
	for _, peer := range c.nodeHub.GetAll() {
		if err := peer.SendMsgLeave(req, &resp); err != nil {
			log.Warnf("peer %s SendMsgLeave failed", peer.Addr())
		}
	}
}

