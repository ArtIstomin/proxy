package handler

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/artistomin/proxy/internal/app/proxy/config"
)

func (p *Proxy) httpsConn(r *http.Request, hostCfg config.Domain,
	cfg *tls.Config) (net.Conn, error) {
	ip := hostCfg.IP
	timeout := time.Duration(hostCfg.Timeout) * time.Second
	dialer := &net.Dialer{
		Timeout: timeout,
	}

	conn, err := tls.DialWithDialer(dialer, "tcp", ip, cfg)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
