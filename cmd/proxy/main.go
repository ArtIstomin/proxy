package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/artistomin/proxy/internal/app/proxy/config"
	"github.com/artistomin/proxy/internal/app/proxy/handlerhttp"
	inmemoryCache "github.com/artistomin/proxy/internal/app/proxy/inmemory"
)

var (
	httpPort  = flag.String("http-port", ":80", "Port for http proxy")
	httpsPort = flag.String("https-port", ":443", "Port for https proxy")
	cfgPath   = flag.String("cfg-path", "configs/config.json", "Path to config file")
)

func main() {
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatal(err)
	}

	storage := inmemoryCache.New()
	transport := initTransport(cfg)
	handlerHTTP := handlerhttp.New(cfg, storage, transport)
	// handlerHTTPS := handlerhttps.New(domainsCfg, storage, transport)

	httpProxy := &http.Server{
		Addr:    *httpPort,
		Handler: handlerHTTP,
	}
	/* httpsProxy := &http.Server{
		Addr:         *httpsPort,
		Handler:      handlerHTTPS,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
	} */

	log.Printf("Listening http on %s and https on %s\n", *httpPort, *httpsPort)
	/* go func() {
		err := httpProxy.ListenAndServe()
		if err != nil {
			log.Fatal(err)
		}
	}() */

	log.Fatal(httpProxy.ListenAndServe())
	// log.Fatal(httpsProxy.ListenAndServeTLS("certs/myCA.cer", "certs/myCA.key"))
}

func initTransport(cfg *config.Config) *http.Transport {
	return &http.Transport{
		MaxConnsPerHost:       cfg.MaxConn,
		IdleConnTimeout:       time.Duration(cfg.IdleTimeout) * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

/* func initPool(domains config.Domains) *connpool.Pool {
	pool := connpool.New()

	for host, cfg := range domains {
		var connFunc connpool.ConnFunc
		if cfg.Pool.Secure {
			connFunc = func(ip string) (net.Conn, error) {
				tlsCfg := certificate.Generate(host)
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
			connFunc = func(ip string) (net.Conn, error) {
				timeout := time.Duration(cfg.Timeout) * time.Second
				dialer := &net.Dialer{
					Timeout:   timeout,
					KeepAlive: timeout,
				}
				conn, err := dialer.Dial("tcp", ip)
				if err != nil {
					return nil, err
				}

				err = conn.(*net.TCPConn).SetKeepAlive(true)

				if err != nil {
					return nil, err
				}

				err = conn.(*net.TCPConn).SetKeepAlivePeriod(timeout)

				if err != nil {
					return nil, err
				}
				return conn, nil
			}
		}

		pool.Init(host, connFunc, cfg.Pool)
	}

	return pool
} */
