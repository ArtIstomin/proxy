package main

import (
	"flag"
	"log"

	"github.com/artistomin/proxy/internal/app/proxy/cache"
	"github.com/artistomin/proxy/internal/app/proxy/config"
	httpproxy "github.com/artistomin/proxy/internal/app/proxy/handlerhttp"
	httpsproxy "github.com/artistomin/proxy/internal/app/proxy/handlerhttps"
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

	storage := cache.New()

	httpProxy := httpproxy.New(domainsCfg, storage, *httpPort)
	httpsProxy := httpsproxy.New(domainsCfg, storage, *httpsPort)

	log.Printf("Listening http on %s and https on %s\n", *httpPort, *httpsPort)
	go func() {
		err := httpProxy.ListenAndServe()
		if err != nil {
			log.Fatal(err)
		}
	}()

	log.Fatal(httpsProxy.ListenAndServeTLS("certs/myCA.cer", "certs/myCA.key"))
}
