package client

import "net"

type Client struct {

	conn            net.Conn

	PeerId          string              //唯一标识

	InvalidPeers    map[string]bool    // 已经无效的peerId
}
