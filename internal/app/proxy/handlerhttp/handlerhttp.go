package handlerhttp

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/artistomin/proxy/internal/app/proxy/cache"
	"github.com/artistomin/proxy/internal/app/proxy/config"
	"github.com/artistomin/proxy/internal/app/proxy/handler"
)

type HttpHandler struct {
	handler.Handler
	tr *http.Transport
}

/* var dialer = &net.Dialer{
	Timeout: 10 * time.Second,
}

// conn, err := hh.GetConn(r)
var tr = &http.Transport{
	Dial: func(network, ip string) (net.Conn, error) {
		ip = "157.150.185.49:80"
		return dialer.Dial(network, ip)
	},
	MaxConnsPerHost: 10,
	IdleConnTimeout: 10 * time.Second,
} */

func (hh *HttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
			log.Printf("panic: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}()

	hostCfg := hh.Config.Domains[r.Host]
	dialer := &net.Dialer{
		Timeout:   time.Duration(hh.Config.Timeout) * time.Second,
		KeepAlive: time.Duration(hh.Config.KeepAlive) * time.Second,
	}
	hh.tr.Dial = func(network, ip string) (net.Conn, error) {
		ip = hostCfg.IP
		return dialer.Dial(network, ip)
	}

	r.URL, _ = url.Parse("http://" + r.Host + r.URL.String())

	hh.RmProxyHeaders(r)
	hh.LogRequest(r, "http")

	if r.Method == http.MethodGet {
		hh.handlerCache(w, r, hostCfg)
	} else {
		hh.handler(w, r)
	}
}

func (hh *HttpHandler) handlerCache(w http.ResponseWriter, r *http.Request, hostCfg config.Domain) {
	host := r.Host
	url := r.URL.String()
	cacheCfg := hostCfg.Cache
	bCacheCfg := hostCfg.BrowserCache

	if bCacheCfg.Enabled {
		ttl := hh.GetTTL(bCacheCfg.TTL, bCacheCfg.TTLUnits)
		ttlStr := strconv.Itoa(ttl)

		w.Header().Set("Cache-Control", "public, max-age="+ttlStr)
	} else {
		w.Header().Del("Cache-Control")
	}

	switch {
	case cacheCfg.Enabled && hh.Cache.Has(host, url):
		cachedValue := hh.Cache.Get(host, url)
		res := &http.Response{
			Status:     cachedValue.Response.Status,
			StatusCode: cachedValue.Response.StatusCode,
			Proto:      cachedValue.Response.Proto,
			ProtoMajor: cachedValue.Response.ProtoMajor,
			ProtoMinor: cachedValue.Response.ProtoMinor,
			Header:     cachedValue.Response.Header,
		}
		bodyReader := bytes.NewReader(cachedValue.Body)

		hh.CopyHeaders(w.Header(), res.Header)

		bytes, err := io.Copy(w, bodyReader)
		if err != nil {
			log.Printf("cache error: %s", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Printf("From cache: %s, bytes: %d", url, bytes)
	default:
		res, err := hh.tr.RoundTrip(r)
		if err != nil {
			log.Printf("request error: %s", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer res.Body.Close()

		bodyReader := io.TeeReader(res.Body, w)

		hh.CopyHeaders(w.Header(), res.Header)

		body, err := ioutil.ReadAll(bodyReader)
		if err != nil {
			log.Printf("body error: %s", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if hh.ShouldResCached(host, r.URL.Path, len(body), cacheCfg) {
			ttl := time.Duration(hh.GetTTL(cacheCfg.TTL, cacheCfg.TTLUnits))
			expireTime := time.Now().UTC().Add(ttl)
			response := cache.Response{
				Status:     res.Status,
				StatusCode: res.StatusCode,
				Proto:      res.Proto,
				ProtoMajor: res.ProtoMajor,
				ProtoMinor: res.ProtoMinor,
				Header:     res.Header,
			}

			hh.Cache.Put(host, url, response, body, expireTime)
		}
	}
}

func (hh *HttpHandler) handler(w http.ResponseWriter, r *http.Request) {
	res, err := hh.tr.RoundTrip(r)
	if err != nil {
		log.Printf("request error: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	hh.CopyHeaders(w.Header(), res.Header)

	bodyReader := io.TeeReader(res.Body, w)

	_, err = ioutil.ReadAll(bodyReader)
	if err != nil {
		log.Printf("body error: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// New creates new http handler
func New(cfg *config.Config, cache cache.Cacher, tr *http.Transport) *HttpHandler {
	return &HttpHandler{handler.Handler{cache, cfg}, tr}
}
