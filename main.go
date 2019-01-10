package main

import (
	"flag"
	"log"
	"net/http"
	"strings"

	"github.com/artistomin/proxy/cache"
	"github.com/artistomin/proxy/handler"
)

var (
	port    = flag.String("port", ":3000", "Port")
	targets = flag.String("targets", "tut.by,mail.ru", "add some sites")
)

func main() {
	flag.Parse()

	targets := *targets
	targetsSlice := strings.Split(targets, ",")

	client := &http.Client{}
	newCache := cache.New()
	proxy := handler.New(client, newCache, *port, targetsSlice)

	log.Printf("Listening http on %s \n", *port)
	log.Fatal(proxy.ListenAndServe())
}
