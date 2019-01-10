package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/artistomin/proxy/cache"
	"github.com/artistomin/proxy/handler"
)

var (
	client  *http.Client
	http200 = []byte("HTTP/1.1 200 Connection Established\r\n\r\n")
	port    = flag.String("Port", ":3000", "Port")
)

func main() {
	flag.Parse()

	client := &http.Client{}
	newCache := cache.New()
	proxy := handler.New(client, newCache, *port)

	log.Printf("Listening http on %s \n", *port)
	log.Fatal(proxy.ListenAndServe())
}
