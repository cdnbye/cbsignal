package client

import (
	"cbsignal/rpcservice"
	"encoding/json"
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

	MAX_NOT_FOUND_PEERS_LIMIT = 5
)

type Client struct {

	Conn            net.Conn

	PeerId          string              //唯一标识

	LocalNode bool             // 是否本地节点

	RpcNodeAddr string       // rpc节点id

	Timestamp int64

	NotFoundPeers     []string   // 记录没有找到的peer的队列
}

type SignalCloseResp struct {
	Action string              `json:"action"`
	FromPeerId string          `json:"from_peer_id,omitempty"`
	Data interface{}           `json:"data,omitempty"`
	Reason string              `json:"reason,omitempty"`
}

type SignalVerResp struct {
	Action string              `json:"action"`
	Ver int                    `json:"ver"`
}

func NewPeerClient(peerId string, conn net.Conn, localNode bool, rpcNodeAddr string) *Client {
	return &Client{
		Conn:        conn,
		PeerId:      peerId,
		LocalNode:   localNode,
		RpcNodeAddr: rpcNodeAddr,
		Timestamp:   time.Now().Unix(),
	}
}

func (c *Client)UpdateTs() {
	//log.Warnf("%s UpdateTs", c.PeerId)
	c.Timestamp = time.Now().Unix()
}

func (c *Client)IsExpired(now, limit int64) bool {
	return now - c.Timestamp > limit
}

func (c *Client)SendMsgClose(reason string) error {
	resp := SignalCloseResp{
		Action: "close",
		Reason: reason,
	}
	b, err := json.Marshal(resp)
	if err != nil {
		log.Error("json.Marshal", err)
		return err
	}
	err, _ = c.SendMessage(b)
	return err
}

func (c *Client)SendMsgVersion(version int) error {
	resp := SignalVerResp{
		Action: "ver",
		Ver: version,
	}
	b, err := json.Marshal(resp)
	if err != nil {
		log.Error("json.Marshal", err)
		return err
	}
	err, _ = c.SendMessage(b)
	return err
}

func (c *Client)SendMessage(msg []byte) (error, bool) {
	return c.sendData(msg, false)
}

func (c *Client)SendBinaryData(data []byte) (error, bool) {
	return c.sendData(data, true)
}

func (c *Client)sendData(data []byte, binary bool) (error, bool) {
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
			return err, true
		}
	} else {
		// 非本地节点
		//log.Warnf("send signal to remote node %s to peer %s", c.RpcNodeAddr, c.PeerId)
		node, ok := rpcservice.GetNode(c.RpcNodeAddr)
		if ok {
			if !node.IsAlive() {
				return fmt.Errorf("node %s is not alive when send signal", node.Addr()), true
			}
			req := rpcservice.SignalReq{
				ToPeerId: c.PeerId,
				Data:     data,
			}
			var resp rpcservice.RpcResp
			err := node.SendMsgSignal(req, &resp)
			if err != nil {
				//log.Warnf("SendMsgSignal to remote failed " + err.Error())
				//log.Warnf("req %+v", req)
				// 节点出现问题
				//node.dialWithChannel()
				return err, true
			}
			if !resp.Success {
				//log.Warnf("SendMsgSignal failed reason " + resp.Reason)
				return fmt.Errorf(resp.Reason), false
			}
		} else {
			log.Warnf("node %s not found", c.RpcNodeAddr)
		}
	}
	return nil, false
}

func (c *Client)Close() error {
	return c.Conn.Close()
}

func (c *Client)EnqueueNotFoundPeer(id string) {
	c.NotFoundPeers = append(c.NotFoundPeers, id)
	if len(c.NotFoundPeers) > MAX_NOT_FOUND_PEERS_LIMIT {
		c.NotFoundPeers = c.NotFoundPeers[1:len(c.NotFoundPeers)]
	}
}

func (c *Client)HasNotFoundPeer(id string) bool {
	for _, v := range c.NotFoundPeers {
		if id == v {
			return true
		}
	}
	return false
}