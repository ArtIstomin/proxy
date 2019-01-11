package handler

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/artistomin/proxy/cache"
)

var hosts = map[string]string{
	"www.tut.by":  "178.172.160.2",
	"www.mail.ru": "94.100.180.201",
}
var http200 = []byte("HTTP/1.1 200 Connection Established\r\n\r\n")

type proxy struct {
	client  *http.Client
	cache   cache.Cache
	targets []string
}

func (p *proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("HERE")
	defer func() {
		if err := recover(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}()

	/* if r.Method == http.MethodConnect {
		p.handlerHTTPS(w, r)
	} else */if r.Method == http.MethodGet {
		fmt.Println("1")
		p.handlerCache(w, r)
	} else {
		p.handlerHTTP(w, r)
	}
}

func (p *proxy) handlerCache(w http.ResponseWriter, r *http.Request) {
	uri := "https://" + hosts[r.Host] + r.URL.String()
	fmt.Println(uri)
	log.Printf("Method: %s, URL: %s", r.Method, uri)

	// cached := p.cache.Has(uri)
	/*
		if cached {
			content := p.cache.Get(uri)
			w.Write(content)
			log.Printf("Bytes: %d, Host: %s, Cached: %t", len(content), r.URL.Host, cached)
			return
		} */

	res, err := p.client.Get(uri)
	// fmt.Printf("%+v\n", res)
	if err != nil {

		fmt.Println("err", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	res.Body.Close()

	w.WriteHeader(res.StatusCode)

	//body, err := ioutil.ReadAll(res.Body)

	// p.cache.Put(uri, body)
	// w.Write(body)

	//log.Printf("Bytes: %d, Host: %s", len(body), r.URL.Host)
}

func (p *proxy) handlerHTTP(w http.ResponseWriter, r *http.Request) {
	rmProxyHeaders(r)
	res, err := p.client.Do(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	copyHeaders(w.Header(), res.Header)
	w.WriteHeader(res.StatusCode)

	_, err = io.Copy(w, res.Body)
	if err != nil && err != io.EOF {
		log.Printf("occur an error when copying remote response to this client, %v", err)
		return
	}
}

/* func (p *proxy) handlerHTTPS(w http.ResponseWriter, r *http.Request) {
	hj, _ := w.(http.Hijacker)
	hjClient, _, err := hj.Hijack()

	if err != nil {
		log.Printf("http user failed to get tcp connection")
		http.Error(w, "Failed", http.StatusBadRequest)
		return
	}

	remote, err := net.Dial("tcp", r.URL.Host)
	if err != nil {
		log.Printf("http user failed connect to '%s'", r.RequestURI)
		http.Error(w, "Failed", http.StatusBadGateway)
		return
	}

	hjClient.Write(http200)

	go copyRemoteToClient(remote, hjClient)
	go copyRemoteToClient(hjClient, remote)
} */

func (p *proxy) containTarget(r *http.Request) bool {
	for _, target := range p.targets {
		if strings.Contains(r.Host, target) || strings.Contains(r.Header.Get("Referer"), target) {
			return true
		}
	}

	return false
}

func copyRemoteToClient(remote, hjClient net.Conn) {
	defer func() {
		remote.Close()
		hjClient.Close()
	}()

	nr, err := io.Copy(remote, hjClient)
	if err != nil && err != io.EOF {
		log.Printf("occur an error when handling CONNECT Method, %v \n", err.Error())
		return
	}

	remoteAddr := remote.RemoteAddr().String()
	clientAddr := hjClient.RemoteAddr().String()
	log.Printf("Bytes: %d, Remote Address: %s, Client Address: %s", nr, remoteAddr, clientAddr)
}

// New creates new proxy server
func New(client *http.Client, cache cache.Cache, port string, targets []string) *http.Server {
	return &http.Server{
		Addr:    port,
		Handler: &proxy{client, cache, targets},
	}
}
