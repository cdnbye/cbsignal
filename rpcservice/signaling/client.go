package signaling

import "cbsignal/rpcservice"

type Client struct {
	nodeHub *rpcservice.NodeHub
}

func NewSignalClient(nodeHub *rpcservice.NodeHub) *Client {
	client := Client{
		nodeHub: nodeHub,
	}
	return &client
}


