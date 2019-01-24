package handlerhttps

import (
	"crypto/tls"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/artistomin/proxy/internal/app/proxy/cache"
	"github.com/artistomin/proxy/internal/app/proxy/certificate"
	"github.com/artistomin/proxy/internal/app/proxy/config"
	"github.com/artistomin/proxy/internal/app/proxy/connpool"
	"github.com/artistomin/proxy/internal/app/proxy/handler"
)

type HttpsHandler struct {
	handler.Handler
}

func (hsh *HttpsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("panic: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}()

	hsh.LogRequest(r, "https")

	hostCfg := hsh.Domains[r.Host]
	if r.Method == http.MethodGet {
		hsh.handlerCache(w, r, hostCfg)
	} else {
		hsh.handler(w, r, hostCfg)
	}
}

func (hsh *HttpsHandler) handlerCache(w http.ResponseWriter, r *http.Request,
	hostCfg config.Domain) {
	var res *http.Response
	var body []byte
	host := r.Host
	url := r.URL.String()
	cacheCfg := hostCfg.Cache
	bCacheCfg := hostCfg.BrowserCache

	switch {
	case cacheCfg.Enabled && hsh.Cache.Has(host, url):
		cachedValue := hsh.Cache.Get(host, url)

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
		conn, err := hsh.httpsConn(r, hostCfg)
		if err != nil {
			log.Printf("connection error: %s", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		res, err = hsh.Request(conn, r)
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

		if hsh.ShouldResCached(host, r.URL.Path, len(body), cacheCfg) {
			ttl := time.Duration(hsh.GetTTL(cacheCfg.TTL, cacheCfg.TTLUnits))
			expireTime := time.Now().UTC().Add(ttl)
			response := cache.Response{
				Status:     res.Status,
				StatusCode: res.StatusCode,
				Proto:      res.Proto,
				ProtoMajor: res.ProtoMajor,
				ProtoMinor: res.ProtoMinor,
				Header:     res.Header,
			}

			hsh.Cache.Put(host, url, response, body, expireTime)
		}
	}

	hsh.CopyHeaders(w.Header(), res.Header)
	w.Header().Del("Date")

	if bCacheCfg.Enabled {
		ttl := hsh.GetTTL(bCacheCfg.TTL, bCacheCfg.TTLUnits)
		ttlStr := strconv.Itoa(ttl)

		w.Header().Set("Cache-Control", "public, max-age="+ttlStr)
	} else {
		w.Header().Del("Cache-Control")
	}

	w.Write(body)
}

func (hsh *HttpsHandler) handler(w http.ResponseWriter, r *http.Request, hostCfg config.Domain) {
	conn, err := hsh.httpsConn(r, hostCfg)
	if err != nil {
		log.Printf("connection error: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	res, err := hsh.Request(conn, r)
	if err != nil {
		log.Printf("request error: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	hsh.CopyHeaders(w.Header(), res.Header)

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("body error: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(body)
}

func (hsh *HttpsHandler) httpsConn(r *http.Request, hostCfg config.Domain) (net.Conn, error) {
	tlsCfg := certificate.Generate(r.Host)
	ip := hostCfg.IP
	timeout := time.Duration(hostCfg.Timeout) * time.Second
	dialer := &net.Dialer{
		Timeout: timeout,
	}

	conn, err := tls.DialWithDialer(dialer, "tcp", ip, tlsCfg)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// New creates new https handler
func New(domains config.Domains, cache cache.Cacher, pool connpool.ConnPool) *HttpsHandler {
	return &HttpsHandler{handler.Handler{cache, domains, pool}}
}
