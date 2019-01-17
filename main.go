package main

import (
	"flag"
	"log"

	"github.com/artistomin/proxy/cache"
	"github.com/artistomin/proxy/config"
	"github.com/artistomin/proxy/handler"
)

var (
	cfgPath = flag.String("cfg-path", "config.json", "Path to config file")
)

type proxyCfg struct {
	port string
	tls  bool
}

func main() {
	flag.Parse()

	httpProxyCfg := proxyCfg{
		port: ":80",
		tls:  false,
	}
	httpsProxyCfg := proxyCfg{
		port: ":443",
		tls:  true,
	}

	domainsCfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatal(err)
	}

	storage := cache.New()
	httpProxy := handler.New(domainsCfg, storage, httpProxyCfg.port, httpProxyCfg.tls)
	httpsProxy := handler.New(domainsCfg, storage, httpsProxyCfg.port, httpsProxyCfg.tls)

	log.Printf("Listening http on %s and https on %s\n", httpProxyCfg.port, httpsProxyCfg.port)
	go func() {
		err := httpProxy.ListenAndServe()
		if err != nil {
			log.Fatal(err)
		}
	}()

	log.Fatal(httpsProxy.ListenAndServeTLS("certs/myCA.cer", "certs/myCA.key"))
}
