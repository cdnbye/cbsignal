package rpcservice

import (
	"github.com/lexkong/log"
	"sync"
)

type NodeHub struct {
	node map[string]*Node
	mu   sync.Mutex
}

var nodeHub *NodeHub

func NewNodeHub() *NodeHub {
	n := NodeHub{
		node: make(map[string]*Node),
	}
	nodeHub = &n
	return &n
}

func GetNode(addr string) (*Node, bool) {
	return nodeHub.Get(addr)
}

func (n *NodeHub) Delete(addr string) {
	log.Warnf("NodeHub delete %s", addr)
	n.mu.Lock()
	delete(n.node, addr)
	n.mu.Unlock()
}

func (n *NodeHub) Add(addr string, peer *Node) {
	log.Infof("NodeHub add %s", addr)
	n.node[addr] = peer
}

func (n *NodeHub) Get(addr string) (*Node, bool) {
	n.mu.Lock()
	peer, ok := n.node[addr]
	n.mu.Unlock()
	return peer, ok
}

func (n *NodeHub) GetAll() map[string]*Node {
	//log.Infof("NodeHub GetAll %d", len(n.node))
	return n.node
}

func (n *NodeHub) Clear() {
	log.Infof("NodeHub clear")
	n.mu.Lock()
	n.node = make(map[string]*Node)
	n.mu.Unlock()
}


