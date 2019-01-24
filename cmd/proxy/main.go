package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/artistomin/proxy/internal/app/proxy/certificate"
	"github.com/artistomin/proxy/internal/app/proxy/config"
	"github.com/artistomin/proxy/internal/app/proxy/connpool"
	"github.com/artistomin/proxy/internal/app/proxy/handlerhttp"
	"github.com/artistomin/proxy/internal/app/proxy/handlerhttps"
	inmemoryCache "github.com/artistomin/proxy/internal/app/proxy/inmemory"
)

var (
	httpPort  = flag.String("http-port", ":80", "Port for http proxy")
	httpsPort = flag.String("https-port", ":443", "Port for https proxy")
	cfgPath   = flag.String("cfg-path", "configs/config.json", "Path to config file")
)

func main() {
	flag.Parse()

	domainsCfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatal(err)
	}

	storage := inmemoryCache.New()
	pool := initPool(domainsCfg)
	fmt.Println(pool)
	handlerHTTP := handlerhttp.New(domainsCfg, storage, pool)
	handlerHTTPS := handlerhttps.New(domainsCfg, storage, pool)

	httpProxy := &http.Server{
		Addr:    *httpPort,
		Handler: handlerHTTP,
	}
	httpsProxy := &http.Server{
		Addr:         *httpsPort,
		Handler:      handlerHTTPS,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
	}

	log.Printf("Listening http on %s and https on %s\n", *httpPort, *httpsPort)
	go func() {
		err := httpProxy.ListenAndServe()
		if err != nil {
			log.Fatal(err)
		}
	}()

	log.Fatal(httpsProxy.ListenAndServeTLS("certs/myCA.cer", "certs/myCA.key"))
}

func initPool(domains config.Domains) connpool.Pool {
	pool := connpool.New()

	for host, cfg := range domains {
		fmt.Println(host)
		var connFunc connpool.ConnFunc
		if cfg.Pool.Secure {
			connFunc = func() (net.Conn, error) {
				tlsCfg := certificate.Generate(host)
				ip := cfg.IP
				timeout := time.Duration(cfg.Timeout) * time.Second
				dialer := &net.Dialer{
					Timeout: timeout,
				}

				conn, err := tls.DialWithDialer(dialer, "tcp", ip, tlsCfg)
				if err != nil {
					return nil, err
				}

				return conn, nil
			}
		} else {
			connFunc = func() (net.Conn, error) {
				ip := cfg.IP
				timeout := time.Duration(cfg.Timeout) * time.Second

				conn, err := net.DialTimeout("tcp", ip, timeout)
				if err != nil {
					return nil, err
				}

				return conn, nil
			}
		}

		pool.Init(host, connFunc, cfg.Pool)
	}

	return pool
}
