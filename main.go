package main

import (
	"flag"
	"log"

	"github.com/artistomin/proxy/cache"
	"github.com/artistomin/proxy/config"
	"github.com/artistomin/proxy/handler"
)

var (
	port    = flag.String("port", ":3000", "Port")
	cfgPath = flag.String("cfg-path", "config.json", "Path to config file")
)

func main() {
	flag.Parse()

	domainsCfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatal(err)
	}

	storage := cache.New()
	proxy := handler.New(domainsCfg, storage, *port)

	log.Printf("Listening http on %s \n", *port)
	log.Fatal(proxy.ListenAndServe())
}

/* client := &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
} */
