package broadcast

import (
	"cbsignal/rpcservice"
	"github.com/lexkong/log"
)

type Client struct {
	Peers map[string]*rpcservice.Peer
}

func NewBroadcastClient(peers map[string]*rpcservice.Peer) *Client {
	client := Client{
		Peers: peers,
	}
	return &client
}

func (c *Client) BroadcastMsgJoin(id string)  {
	req := JoinLeaveReq{
		Id: id,
	}
	var resp JoinLeaveResp
	for _, peer := range c.Peers {
		if err := peer.SendMsgJoin(req, &resp); err != nil {
			log.Warnf("peer %s SendMsgJoin failed", peer.Addr())
		}
	}
}

func (c *Client) BroadcastMsgLeave(id string)  {
	req := JoinLeaveReq{
		Id: id,
	}
	var resp JoinLeaveResp
	for _, peer := range c.Peers {
		if err := peer.SendMsgLeave(req, &resp); err != nil {
			log.Warnf("peer %s SendMsgLeave failed", peer.Addr())
		}
	}
}

