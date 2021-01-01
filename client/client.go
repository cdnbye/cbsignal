package client

import (
	"cbsignal/rpcservice"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/lexkong/log"
	"net"
	"time"
)

const (

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

)

type Client struct {

	Conn            net.Conn

	PeerId          string              //唯一标识

	InvalidPeers    map[string]bool    // 已经无效的peerId

	CompressSupported bool             // 是否支持压缩

	LocalNode bool             // 是否本节点

	RpcNodeAddr string       // rpc节点id
}

func (c *Client)SendMessage(msg []byte) error {
	return c.sendData(msg, false)
}

func (c *Client)SendBinaryData(data []byte) error {
	return c.sendData(data, true)
}

func (c *Client)sendData(data []byte, binary bool) error {
	var opCode ws.OpCode
	if binary {
		opCode = ws.OpBinary
	} else {
		opCode = ws.OpText
	}
	//log.Infof("client send data %t", c.LocalNode)
	if c.LocalNode {
		// 本地节点
		err := wsutil.WriteServerMessage(c.Conn, opCode, data)
		if err != nil {
			// handle error
			log.Warnf("WriteServerMessage " + err.Error())
			return err
		}
	} else {
		// 非本地节点
		//log.Infof("send data to addr %s", c.RpcNodeAddr)
		node, ok := rpcservice.GetNode(c.RpcNodeAddr)
		if ok {
			req := rpcservice.SignalReq{
				ToPeerId: c.PeerId,
				Data:     data,
			}
			var resp rpcservice.RpcResp
			err := node.SendMsgSignal(req, &resp)
			if err != nil {
				log.Warnf("SendMsgSignal " + err.Error())
				return err
			}
			if !resp.Success {
				log.Warnf("SendMsgSignal failed reason " + resp.Reason)
			}
		} else {
			log.Warnf("node %s not found", c.RpcNodeAddr)
		}
	}
	return nil
}