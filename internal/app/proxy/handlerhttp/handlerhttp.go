package handlerhttp

import (
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/artistomin/proxy/internal/app/proxy/cache"
	"github.com/artistomin/proxy/internal/app/proxy/config"
	"github.com/artistomin/proxy/internal/app/proxy/handler"
)

type httpProxy struct {
	handler.Proxy
}

func (hp *httpProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("panic: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}()

	hp.LogRequest(r, "http")

	hostCfg := hp.Domains[r.Host]
	if r.Method == http.MethodGet {
		hp.handlerCache(w, r, hostCfg)
	} else {
		hp.handler(w, r, hostCfg)
	}
}

func (hp *httpProxy) handlerCache(w http.ResponseWriter, r *http.Request, hostCfg config.Domain) {
	var res *http.Response
	var body []byte
	host := r.Host
	url := r.URL.String()
	cacheCfg := hostCfg.Cache
	bCacheCfg := hostCfg.BrowserCache

	switch {
	case cacheCfg.Enabled && hp.Cache.Has(host, url):
		cachedValue := hp.Cache.Get(host, url)

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
	default:
		conn, err := hp.httpConn(r, hostCfg)
		if err != nil {
			log.Printf("connection error: %s", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		res, err = hp.Request(conn, r)
		if err != nil {
			log.Printf("request error: %s", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer res.Body.Close()

		body, err = ioutil.ReadAll(res.Body)
		if err != nil {
			log.Printf("body error: %s", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if hp.ShouldResCached(host, r.URL.Path, len(body), cacheCfg) {
			ttl := time.Duration(hp.GetTTL(cacheCfg.TTL, cacheCfg.TTLUnits))
			expireTime := time.Now().UTC().Add(ttl)
			response := cache.Response{
				Status:     res.Status,
				StatusCode: res.StatusCode,
				Proto:      res.Proto,
				ProtoMajor: res.ProtoMajor,
				ProtoMinor: res.ProtoMinor,
				Header:     res.Header,
			}

			hp.Cache.Put(host, url, response, body, expireTime)
		}
	}

	hp.CopyHeaders(w.Header(), res.Header)
	w.Header().Del("Date")

	if bCacheCfg.Enabled {
		ttl := hp.GetTTL(bCacheCfg.TTL, bCacheCfg.TTLUnits)
		ttlStr := strconv.Itoa(ttl)

		w.Header().Set("Cache-Control", "public, max-age="+ttlStr)
	} else {
		w.Header().Del("Cache-Control")
	}

	w.Write(body)
}

func (hp *httpProxy) handler(w http.ResponseWriter, r *http.Request, hostCfg config.Domain) {
	conn, err := hp.httpConn(r, hostCfg)
	if err != nil {
		log.Printf("connection error: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	res, err := hp.Request(conn, r)
	if err != nil {
		log.Printf("request error: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	hp.CopyHeaders(w.Header(), res.Header)

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("body error: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(body)
}

func (hp *httpProxy) httpConn(r *http.Request, hostCfg config.Domain) (net.Conn, error) {
	ip := hostCfg.IP
	timeout := time.Duration(hostCfg.Timeout) * time.Second

	conn, err := net.DialTimeout("tcp", ip, timeout)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// New creates new http proxy server
func New(domains config.Domains, cache cache.Cacher, port string) *http.Server {
	return &http.Server{
		Addr:    port,
		Handler: &httpProxy{handler.Proxy{cache, domains}},
	}
}
