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
	for _, node := range c.nodeHub.GetAll() {
		if err := node.SendMsgJoin(req, &resp); err != nil {
			log.Warnf("node %s SendMsgJoin failed", node.Addr())
		}
	}
}

func (c *Client) BroadcastMsgLeave(id string)  {
	req := rpcservice.JoinLeaveReq{
		PeerId: id,
	}
	var resp rpcservice.RpcResp
	for _, node := range c.nodeHub.GetAll() {
		if err := node.SendMsgLeave(req, &resp); err != nil {
			log.Warnf("node %s SendMsgLeave failed", node.Addr())
		}
	}
}

