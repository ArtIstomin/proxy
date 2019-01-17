package handler

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/artistomin/proxy/config"
)

func (p *proxy) httpConn(r *http.Request, hostCfg config.Domain) (net.Conn, error) {
	ip := hostCfg.IP
	timeout := time.Duration(hostCfg.Timeout) * time.Second

	conn, err := net.DialTimeout("tcp", ip, timeout)
	if err != nil {
		return nil, fmt.Errorf("TCP connection error: %s", err)
	}

	return conn, nil
}
