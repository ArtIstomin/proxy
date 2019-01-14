package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"

	"github.com/artistomin/proxy/cache"
	"github.com/artistomin/proxy/handler"
)

var (
	port = flag.String("port", ":3000", "Port")
)

func main() {
	flag.Parse()

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	newCache := cache.New()
	proxy := handler.New(client, newCache, *port)

	log.Printf("Listening http on %s \n", *port)
	log.Fatal(proxy.ListenAndServe())
}
