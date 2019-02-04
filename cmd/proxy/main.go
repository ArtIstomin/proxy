package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/artistomin/proxy/internal/app/proxy/config"
	"github.com/artistomin/proxy/internal/app/proxy/handlerhttp"
	"github.com/artistomin/proxy/internal/app/proxy/handlerhttps"
	"github.com/artistomin/proxy/internal/app/proxy/inmemory"
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

	storage := inmemory.New()
	transport := initTransport(cfg)
	handlerHTTP := handlerhttp.New(cfg, storage, transport)
	handlerHTTPS := handlerhttps.New(cfg, storage, transport)

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

func initTransport(cfg *config.Config) *http.Transport {
	return &http.Transport{
		MaxConnsPerHost:       cfg.MaxConn,
		IdleConnTimeout:       cfg.IdleTimeout * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}
