package handler

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/artistomin/proxy/config"
)

func (p *proxy) httpsConn(r *http.Request, hostCfg config.Domain,
	cfg *tls.Config) (net.Conn, error) {
	ip := hostCfg.IP
	timeout := time.Duration(hostCfg.Timeout) * time.Second
	dialer := &net.Dialer{
		Timeout: timeout,
	}

	conn, err := tls.DialWithDialer(dialer, "tcp", ip, cfg)
	if err != nil {
		return nil, fmt.Errorf("TLS connection error: %s", err)
	}

	return conn, nil
}
