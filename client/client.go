package client

import (
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
}

func (c *Client)SendMessage(msg []byte) error {
	err := wsutil.WriteServerMessage(c.Conn, ws.OpText, msg)
	if err != nil {
		// handle error
		log.Warnf("WriteServerMessage " + err.Error())
		return err
	}
	return nil
}
