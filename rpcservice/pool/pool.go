package pool

import (
	"errors"
	"net/rpc"
)

var (
	//ErrClosed 连接池已经关闭Error
	ErrClosed = errors.New("pool is closed")
)

type Pool interface {
	Acquire() (*rpc.Client, error) // 获取资源
	Release(*rpc.Client) error     // 释放资源
	Close(*rpc.Client) error       // 关闭资源
	Shutdown() error             // 关闭池
	NumTotalConn() int
	NumIdleConn() int
}
