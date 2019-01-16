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
	"strings"
	"time"

	"github.com/artistomin/proxy/cache"
	"github.com/artistomin/proxy/config"
)

const sizeValue = 1024

type proxy struct {
	cache   *cache.Cache
	domains config.Domains
}

func (p *proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Panic: %v", err)
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
	var res *http.Response
	var body []byte
	host := r.Host
	url := r.URL.String()
	cacheCfg := hostCfg.Cache
	bCacheCfg := hostCfg.BrowserCache

	if cacheCfg.Enabled && p.cache.Has(host, url) {
		cachedValue := p.cache.Get(host, url)

		body = cachedValue.Body
		res = &http.Response{
			Status:     cachedValue.Response.Status,
			StatusCode: cachedValue.Response.StatusCode,
			Proto:      cachedValue.Response.Proto,
			ProtoMajor: cachedValue.Response.ProtoMajor,
			ProtoMinor: cachedValue.Response.ProtoMinor,
			Header:     cachedValue.Response.Header,
		}

		log.Printf("From cache: %s, Bytes: %d", url, len(body))
	} else {
		conn, err := p.tcpConn(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		res, err = p.HTTPRequest(conn, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer res.Body.Close()

		body, err = ioutil.ReadAll(res.Body)
		if err != nil {
			log.Printf("Body error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if p.shouldResCached(host, r.URL.Path, len(body), cacheCfg) {
			ttl := time.Duration(getTTL(cacheCfg.TTL, cacheCfg.TTLUnits))
			expireTime := time.Now().UTC().Add(ttl)
			response := cache.Response{
				Status:     res.Status,
				StatusCode: res.StatusCode,
				Proto:      res.Proto,
				ProtoMajor: res.ProtoMajor,
				ProtoMinor: res.ProtoMinor,
				Header:     res.Header,
			}

			p.cache.Put(host, url, response, body, expireTime)
		}
	}

	copyHeaders(w.Header(), res.Header)
	w.Header().Del("Date")

	if bCacheCfg.Enabled {
		ttl := getTTL(bCacheCfg.TTL, bCacheCfg.TTLUnits)
		ttlStr := strconv.Itoa(ttl)

		w.Header().Set("Cache-Control", "public, max-age="+ttlStr)
	} else {
		w.Header().Del("Cache-control")
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
		log.Printf("Body error: %v", err)
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

func (p *proxy) shouldResCached(host, path string, bodySize int, cacheCfg config.Cache) bool {
	if !cacheCfg.Enabled {
		return false
	}

	if (p.cache.Size(host) + bodySize) >= maxSizeBytes(cacheCfg.MaxSize, cacheCfg.SizeUnits) {
		return false
	}

	if bodySize > maxSizeBytes(cacheCfg.CacheObject.MaxSize, cacheCfg.CacheObject.SizeUnits) {
		return false
	}

	if len(cacheCfg.Cached) > 0 && !pathHasSuffix(path, cacheCfg.Cached) {
		return false
	}

	if len(cacheCfg.NoCached) > 0 && pathContainsString(path, cacheCfg.NoCached) {
		return false
	}

	return true
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

func maxSizeBytes(size int, units string) int {
	var maxSize int
	units = strings.ToLower(units)

	switch units {
	case "kb":
		maxSize = size * sizeValue
	case "mb":
		maxSize = size * sizeValue * sizeValue
	case "gb":
		maxSize = size * sizeValue * sizeValue * sizeValue
	}

	return maxSize
}

func pathHasSuffix(path string, suffixes []string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(path, suffix) {
			return true
		}
	}

	return false
}

func pathContainsString(path string, subStrings []string) bool {
	for _, subString := range subStrings {
		if strings.Contains(path, subString) {
			return true
		}
	}

	return false
}

// New creates new proxy server
func New(domains config.Domains, cache *cache.Cache, port string) *http.Server {
	return &http.Server{
		Addr:    port,
		Handler: &proxy{cache, domains},
	}
}
