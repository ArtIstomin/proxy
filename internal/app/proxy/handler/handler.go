package handler

import (
	"bufio"
	"crypto/tls"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"strconv"
	"time"

	"github.com/artistomin/proxy/internal/app/proxy/cache"
	"github.com/artistomin/proxy/internal/app/proxy/certificate"
	"github.com/artistomin/proxy/internal/app/proxy/config"
)

const sizeValue = 1024

type Proxy struct {
	cache      cache.Cacher
	domains    config.Domains
	tlsEnabled bool
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("panic: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}()

	if p.tlsEnabled {
		logRequest(r, "https")
	} else {
		logRequest(r, "http")
	}

	hostCfg := p.domains[r.Host]
	if r.Method == http.MethodGet {
		p.handlerCache(w, r, hostCfg)
	} else {
		p.handler(w, r, hostCfg)
	}
}

func (p *Proxy) handlerCache(w http.ResponseWriter, r *http.Request, hostCfg config.Domain) {
	var res *http.Response
	var body []byte
	host := r.Host
	url := r.URL.String()
	cacheCfg := hostCfg.Cache
	bCacheCfg := hostCfg.BrowserCache

	switch {
	case cacheCfg.Enabled && p.cache.Has(host, url):
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
	default:
		var conn net.Conn
		var err error

		if p.tlsEnabled {
			tlsCfg := certificate.Generate(r)
			conn, err = p.httpsConn(r, hostCfg, tlsCfg)
		} else {
			conn, err = p.httpConn(r, hostCfg)
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		res, err = p.Request(conn, r)
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
		w.Header().Del("Cache-Control")
	}

	w.Write(body)
}

func (p *Proxy) handler(w http.ResponseWriter, r *http.Request, hostCfg config.Domain) {
	var conn net.Conn
	var err error

	if p.tlsEnabled {
		tlsCfg := certificate.Generate(r)
		conn, err = p.httpsConn(r, hostCfg, tlsCfg)
	} else {
		conn, err = p.httpConn(r, hostCfg)
	}

	if err != nil {
		log.Printf("Request error: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	res, err := p.Request(conn, r)
	if err != nil {
		log.Printf("Body error: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	copyHeaders(w.Header(), res.Header)

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("Body error: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(body)
}

// generic
func (p *Proxy) Request(conn net.Conn, r *http.Request) (*http.Response, error) {
	rmProxyHeaders(r)

	dumpReq, err := httputil.DumpRequest(r, true)
	if err != nil {
		return nil, err
	}

	_, err = conn.Write(dumpReq)
	if err != nil {
		return nil, err
	}

	resReader := bufio.NewReader(conn)
	if err != nil {
		return nil, err
	}

	res, err := http.ReadResponse(resReader, r)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// generic
func (p *Proxy) shouldResCached(host, path string, bodySize int, cacheCfg config.Cache) bool {
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

// New creates new proxy server
func New(domains config.Domains, cache cache.Cacher, port string, tlsEnabled bool) *http.Server {
	var server *http.Server

	if tlsEnabled {
		server = &http.Server{
			Addr:         port,
			Handler:      &Proxy{cache, domains, tlsEnabled},
			TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
		}
	} else {
		server = &http.Server{
			Addr:    port,
			Handler: &Proxy{cache, domains, tlsEnabled},
		}
	}

	return server
}
