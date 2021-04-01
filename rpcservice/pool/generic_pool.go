package pool

import (
	"errors"
	"net/rpc"
	"sync"
	"time"
)

var (
	ErrInvalidConfig = errors.New("invalid pool config")
	ErrPoolClosed    = errors.New("pool closed")
)

type factory func() (*rpc.Client, error)



type GenericPool struct {
	sync.Mutex
	pool        chan *rpc.Client
	maxOpen     int  // 池中最大资源数
	numOpen     int  // 当前池中资源数
	minOpen     int  // 池中最少资源数
	closed      bool // 池是否已关闭
	maxLifetime time.Duration
	factory     factory // 创建连接的方法
}

func NewGenericPool(minOpen, maxOpen int, maxLifetime time.Duration, factory factory) (*GenericPool, error) {
	if maxOpen <= 0 || minOpen > maxOpen {
		return nil, ErrInvalidConfig
	}
	p := &GenericPool{
		maxOpen:     maxOpen,
		minOpen:     minOpen,
		maxLifetime: maxLifetime,
		factory:     factory,
		pool:        make(chan *rpc.Client, maxOpen),
	}

	for i := 0; i < minOpen; i++ {
		closer, err := factory()
		if err != nil {
			continue
		}
		p.numOpen++
		p.pool <- closer
	}
	return p, nil
}

func (p *GenericPool) Acquire() (*rpc.Client, error) {
	if p.closed {
		return nil, ErrPoolClosed
	}
	for {
		closer, err := p.getOrCreate()
		if err != nil {
			return nil, err
		}
		// todo maxLifttime处理
		return closer, nil
	}
}

func (p *GenericPool) getOrCreate() (*rpc.Client, error) {
	select {
	case closer := <-p.pool:
		return closer, nil
	default:
	}
	p.Lock()
	if p.numOpen < p.maxOpen {
		// 新建连接
		closer, err := p.factory()
		if err != nil {
			p.Unlock()
			return nil, err
		}
		p.numOpen++
		p.Unlock()
		return closer, nil
	}
	p.Unlock()
	closer := <-p.pool
	return closer, nil

	//if p.numOpen >= p.maxOpen {
	//	//p.Unlock()
	//	closer := <-p.pool
	//	return closer, nil
	//}
	// 新建连接
	//closer, err := p.factory()
	//if err != nil {
	//	//p.Unlock()
	//	return nil, err
	//}
	//p.Lock()
	//p.numOpen++
	//p.Unlock()
	//return closer, nil
}

// 释放单个资源到连接池
func (p *GenericPool) Release(closer *rpc.Client) error {
	if p.closed {
		return ErrPoolClosed
	}
	//log.Warnf("Release conn now %d idle %d", p.NumTotalConn(), p.NumIdleConn())
	p.Lock()
	p.pool <- closer
	p.Unlock()
	return nil
}



// 关闭单个资源
func (p *GenericPool) Close(closer *rpc.Client) error {
	p.Lock()
	closer.Close()
	p.numOpen--
	p.Unlock()
	return nil
}

func (p *GenericPool) NumTotalConn() int {
	return p.numOpen
}

func (p *GenericPool) NumIdleConn() int {
	return len(p.pool)
}


// 关闭连接池，释放所有资源
func (p *GenericPool) Shutdown() error {
	if p.closed {
		return ErrPoolClosed
	}
	p.Lock()
	close(p.pool)
	for closer := range p.pool {
		closer.Close()
		p.numOpen--
	}
	p.closed = true
	p.Unlock()
	return nil
}
