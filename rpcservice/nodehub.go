package rpcservice

import (
	"github.com/lexkong/log"
	"sync"
)

type NodeHub struct {
	peers map[string]*Node
	mu sync.Mutex
}

var nodeHub *NodeHub

func NewNodeHub() *NodeHub {
	n := NodeHub{
		peers: make(map[string]*Node),
	}
	nodeHub = &n
	return &n
}

func GetNode(addr string) (*Node, bool) {
	return nodeHub.Get(addr)
}

func (n *NodeHub) Delete(addr string) {
	log.Infof("NodeHub delete %s", addr)
	n.mu.Lock()
	delete(n.peers, addr)
	n.mu.Unlock()
}

func (n *NodeHub) Add(addr string, peer *Node) {
	log.Infof("NodeHub add %s", addr)
	n.peers[addr] = peer
}

func (n *NodeHub) Get(addr string) (*Node, bool) {
	n.mu.Lock()
	peer, ok := n.peers[addr]
	n.mu.Unlock()
	return peer, ok
}

func (n *NodeHub) GetAll() map[string]*Node {
	//log.Infof("NodeHub GetAll %d", len(n.peers))
	return n.peers
}

func (n *NodeHub) Clear() {
	log.Infof("NodeHub clear")
	n.mu.Lock()
	n.peers = make(map[string]*Node)
	n.mu.Unlock()
}


