package rpcservice

import (
	"github.com/lexkong/log"
	"sync"
)

type NodeHub struct {
	nodes map[string]*Node
	mu   sync.Mutex
}

var nodeHub *NodeHub

func NewNodeHub() *NodeHub {
	n := NodeHub{
		nodes: make(map[string]*Node),
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
	if node, ok := n.nodes[addr]; ok {
		node.Released = true
		node.connPool.Shutdown()
	}
	delete(n.nodes, addr)
	n.mu.Unlock()
}

func (n *NodeHub) Add(addr string, peer *Node) {
	log.Infof("NodeHub add %s", addr)
	n.nodes[addr] = peer
}

func (n *NodeHub) Get(addr string) (*Node, bool) {
	n.mu.Lock()
	node, ok := n.nodes[addr]
	n.mu.Unlock()
	return node, ok
}

func (n *NodeHub) GetAll() map[string]*Node {
	//log.Infof("NodeHub GetAll %d", len(n.node))
	return n.nodes
}

func (n *NodeHub) Clear() {
	log.Infof("NodeHub clear")
	n.mu.Lock()
	for _, node := range n.nodes {
		node.connPool.Shutdown()
	}
	n.nodes = make(map[string]*Node)
	n.mu.Unlock()
}


