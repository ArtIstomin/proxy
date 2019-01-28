package connpool

import (
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/artistomin/proxy/internal/app/proxy/config"
)

type ConnFunc func(ip string) (net.Conn, error)

type ConnPool interface {
	Init(host string, connFunc ConnFunc, poolCfg config.Pool)
	Get(host, ip string) (net.Conn, error)
	Put(host string, conn net.Conn) error
}

type Host struct {
	conns      chan net.Conn
	maxConn    int
	totalConn  int
	createConn ConnFunc
}

type HostPool map[string]*Host

type Pool struct {
	*sync.Mutex
	HostPool
}

func (p *Pool) Get(host string, ip string) (net.Conn, error) {
	hostPool := p.HostPool[host]
	go func() {
		conn, err := p.createConn(host, ip)
		if err != nil {
			return
		}

		hostPool.conns <- conn
	}()

	select {
	case conn := <-hostPool.conns:
		return conn, nil
	}
}

func (p *Pool) Put(host string, conn net.Conn) error {
	if conn == nil {
		p.Lock()
		p.HostPool[host].totalConn--
		p.Unlock()
		return errors.New("Cannot put nil to connection pool.")
	}

	select {
	case p.HostPool[host].conns <- conn:
		return nil
	default:
		return conn.Close()
	}
}

func (p *Pool) Init(host string, connFunc ConnFunc, poolCfg config.Pool) {
	v := &Host{}
	v.maxConn = poolCfg.MaxConn
	v.conns = make(chan net.Conn, v.maxConn)
	v.createConn = connFunc

	p.HostPool[host] = v
}

func (p *Pool) createConn(host, ip string) (net.Conn, error) {
	p.Lock()
	defer p.Unlock()

	hostPool := p.HostPool[host]

	if hostPool.totalConn >= hostPool.maxConn {
		return nil, fmt.Errorf("pool error: maximum connections exceeded")
	}

	conn, err := hostPool.createConn(ip)
	if err != nil {
		return nil, fmt.Errorf("pool erorr: cannot create new connection: %s", err)
	}
	hostPool.totalConn++

	return conn, nil
}

func New() *Pool {
	return &Pool{
		&sync.Mutex{},
		make(HostPool),
	}
}
