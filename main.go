package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"strings"

	"github.com/artistomin/proxy/cache"
	"github.com/artistomin/proxy/handler"
)

var (
	targets = flag.String("targets", "tut.by,mail.ru", "add some sites")
)

func main() {
	flag.Parse()

	httpPort := ":80"
	httpsPort := ":443"

	targets := *targets
	targetsSlice := strings.Split(targets, ",")

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	newCache := cache.New()
	//proxy := handler.New(client, newCache, httpPort, targetsSlice)
	proxyTLS := handler.New(client, newCache, httpsPort, targetsSlice)

	log.Printf("Listening http on %s \n", httpPort)
	log.Printf("Listening https on %s \n", httpsPort)

	// go proxy.ListenAndServe()
	log.Fatal(proxyTLS.ListenAndServeTLS("server.pem", "server.key"))
}
