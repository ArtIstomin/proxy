package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/artistomin/proxy/internal/app/proxy/cache"
	"github.com/artistomin/proxy/internal/app/proxy/config"
	"github.com/artistomin/proxy/internal/app/proxy/grpcclient"
	"github.com/artistomin/proxy/internal/app/proxy/handlerhttp"
	"github.com/artistomin/proxy/internal/app/proxy/handlerhttps"
)

var (
	httpPort    = flag.String("http-port", ":80", "Port for http proxy")
	httpsPort   = flag.String("https-port", ":443", "Port for https proxy")
	grpcAddress = flag.String("grpc-port", ":3000", "Port for grpc client")
	cfgPath     = flag.String("cfg-path", "configs/config.json", "Path to config file")
	crtPath     = flag.String("crt-path", "certs", "Path to certificates")
	inmemory    = flag.Bool("inmemory-cache", false, "Use in memory cache")
)

func main() {
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatal(err)
	}

	grpcConn, grpcClient, err := grpcclient.NewActivityClient(*grpcAddress)
	if err != nil {
		log.Fatalf("did not connect to grpc: %s", err)
	}
	defer grpcConn.Close()

	var storage cache.Cacher
	if *inmemory {
		storage = cache.NewInmemory()
	} else {
		storage = cache.NewActivity(grpcClient)
	}

	transport := initTransport(cfg)
	handlerHTTP := handlerhttp.New(cfg, storage, transport)
	handlerHTTPS := handlerhttps.New(cfg, storage, transport)

	httpProxy := &http.Server{
		Addr:    *httpPort,
		Handler: handlerHTTP,
	}
	httpsProxy := &http.Server{
		Addr:    *httpsPort,
		Handler: handlerHTTPS,
	}

	log.Printf("Listening http on %s and https on %s\n", *httpPort, *httpsPort)
	go func() {
		err := httpProxy.ListenAndServe()
		if err != nil {
			log.Fatal(err)
		}
	}()

	log.Fatal(httpsProxy.ListenAndServeTLS(*crtPath+"/myCA.cer", *crtPath+"/myCA.key"))
}

func initTransport(cfg *config.Config) *http.Transport {
	return &http.Transport{
		MaxConnsPerHost:       cfg.MaxConn,
		IdleConnTimeout:       cfg.IdleTimeout * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}
