package connpool

import (
	"fmt"
	"net"
	"sync"

	"github.com/artistomin/proxy/internal/app/proxy/config"
)

type ConnFunc func() (net.Conn, error)

type ConnPool interface {
	Init(host string, connFunc ConnFunc, poolCfg config.Pool)
	Get(host string) (net.Conn, error)
	Remove(conn net.Conn) error
}

type Host struct {
	conns      chan net.Conn
	maxConn    int
	freeConn   int
	createConn ConnFunc
}

type HostPool map[string]Host

type Pool struct {
	*sync.Mutex
	HostPool
}

func (p Pool) Get(host string) (net.Conn, error) {
	p.Lock()
	defer p.Unlock()
	fmt.Println("\\\\\\\\\\\\\\\\\\\\\\\\\\")
	fmt.Println(host)
	fmt.Println(p.HostPool[host])
	return p.HostPool[host].createConn()
}

func (p Pool) Remove(conn net.Conn) error {
	return nil
}

func (p Pool) Init(host string, connFunc ConnFunc, poolCfg config.Pool) {
	v := Host{}
	v.freeConn = 0
	v.maxConn = poolCfg.MaxConn
	v.conns = make(chan net.Conn, v.maxConn)
	v.createConn = connFunc

	p.HostPool[host] = v
}

func New() Pool {
	return Pool{
		&sync.Mutex{},
		make(HostPool),
	}
}

func isClosedConn(conn net.Conn) bool {
	return false
}
