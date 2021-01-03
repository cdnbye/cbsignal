package client

import (
	"cbsignal/rpcservice"
	"fmt"
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
			log.Infof("WriteServerMessage " + err.Error())
			return err
		}
	} else {
		// 非本地节点
		//log.Warnf("send signal to remote node %s to peer %s", c.RpcNodeAddr, c.PeerId)
		node, ok := rpcservice.GetNode(c.RpcNodeAddr)
		if ok {
			req := rpcservice.SignalReq{
				ToPeerId: c.PeerId,
				Data:     data,
			}
			var resp rpcservice.RpcResp
			err := node.SendMsgSignal(req, &resp)
			if err != nil {
				log.Warnf("SendMsgSignal to remote failed " + err.Error())
				// 节点出现问题
				return err
			}
			if !resp.Success {
				log.Warnf("SendMsgSignal failed reason " + resp.Reason)
				return fmt.Errorf(resp.Reason)
			}
		} else {
			log.Warnf("node %s not found", c.RpcNodeAddr)
		}
	}
	return nil
}