package handler

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"strconv"
	"time"

	"github.com/artistomin/proxy/cache"
	"github.com/artistomin/proxy/config"
)

type proxy struct {
	cache   *cache.Cache
	domains config.Domains
}

func (p *proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}()

	log.Printf("Method: %s, Host: %s, Url: %s\n", r.Method, r.Host, r.URL.String())
	hostCfg := p.domains[r.Host]

	if r.Method != http.MethodGet {
		p.handlerHTTP(w, r)
	} else {
		p.handlerCache(w, r, hostCfg)
	}
}

func (p *proxy) handlerCache(w http.ResponseWriter, r *http.Request, hostCfg config.Domain) {
	// cache := hostCfg.Cache
	bCache := hostCfg.BrowserCache

	conn, err := p.tcpConn(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	res, err := p.HTTPRequest(conn, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	copyHeaders(w.Header(), res.Header)

	if bCache.Enabled {
		ttl := getTTL(bCache.TTL, bCache.TTLUnits)
		ttlStr := strconv.Itoa(ttl)

		w.Header().Set("Cache-Control", "public, max-age="+ttlStr)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(body)
}

func (p *proxy) handlerHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := p.tcpConn(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	res, err := p.HTTPRequest(conn, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	copyHeaders(w.Header(), res.Header)

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(body)
}

func (p *proxy) HTTPRequest(conn net.Conn, r *http.Request) (*http.Response, error) {
	rmProxyHeaders(r)

	dumpReq, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Printf("TCP error: %v\n", err.Error())
		return nil, err
	}

	_, err = conn.Write(dumpReq)
	if err != nil {
		log.Printf("TCP error: %v\n", err.Error())
		return nil, err
	}

	resReader := bufio.NewReader(conn)
	if err != nil {
		log.Printf("TCP error: %v\n", err.Error())
		return nil, err
	}

	res, err := http.ReadResponse(resReader, r)
	if err != nil {
		log.Printf("TCP error: %v\n", err.Error())
		return nil, err
	}

	return res, nil
}

func (p *proxy) tcpConn(r *http.Request) (net.Conn, error) {
	ip := p.domains[r.Host].IP
	timeout := time.Duration(p.domains[r.Host].Timeout) * time.Second

	conn, err := net.DialTimeout("tcp", ip, timeout)
	if err != nil {
		return nil, fmt.Errorf("TCP connection error: %s", err)
	}

	return conn, nil
}

func getTTL(ttl time.Duration, units string) int {
	var ttlDuration time.Duration

	switch units {
	case "h":
		ttlDuration = ttl * time.Hour
	case "s":
		ttlDuration = ttl * time.Second
	case "m":
		ttlDuration = ttl * time.Minute
	}

	return int(ttlDuration.Seconds())
}

func filterRequest(URL string) bool {
	/* for _, extension := range extensions {
		if strings.Contains(URL, extension) {
			return true
		}
	} */

	return false
}

// New creates new proxy server
func New(domains config.Domains, cache *cache.Cache, port string) *http.Server {
	return &http.Server{
		Addr:    port,
		Handler: &proxy{cache, domains},
	}
}

/* if hostCfg.Cache.Enabled && hostCfg.BrowserCache.Enabled {
	fmt.Println("Cache and browser cache enabled")
} else if hostCfg.Cache.Enabled {
	fmt.Println("Cache enabled")
} else if hostCfg.BrowserCache.Enabled {
	fmt.Println("Browser cache enabled")
} else {
	fmt.Println("Cache and browser cache disabled")
	p.handlerHTTP(w, r)
} */
