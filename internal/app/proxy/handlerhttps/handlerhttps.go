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
	"github.com/artistomin/proxy/internal/app/proxy/handler"
)

type httpsProxy struct {
	handler.Proxy
}

func (hsp *httpsProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("panic: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}()

	hsp.LogRequest(r, "https")

	hostCfg := hsp.Domains[r.Host]
	if r.Method == http.MethodGet {
		hsp.handlerCache(w, r, hostCfg)
	} else {
		hsp.handler(w, r, hostCfg)
	}
}

func (hsp *httpsProxy) handlerCache(w http.ResponseWriter, r *http.Request, hostCfg config.Domain) {
	var res *http.Response
	var body []byte
	host := r.Host
	url := r.URL.String()
	cacheCfg := hostCfg.Cache
	bCacheCfg := hostCfg.BrowserCache

	switch {
	case cacheCfg.Enabled && hsp.Cache.Has(host, url):
		cachedValue := hsp.Cache.Get(host, url)

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
		tlsCfg := certificate.Generate(r)
		conn, err := hsp.httpsConn(r, hostCfg, tlsCfg)
		if err != nil {
			log.Printf("connection error: %s", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		res, err = hsp.Request(conn, r)
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

		if hsp.ShouldResCached(host, r.URL.Path, len(body), cacheCfg) {
			ttl := time.Duration(hsp.GetTTL(cacheCfg.TTL, cacheCfg.TTLUnits))
			expireTime := time.Now().UTC().Add(ttl)
			response := cache.Response{
				Status:     res.Status,
				StatusCode: res.StatusCode,
				Proto:      res.Proto,
				ProtoMajor: res.ProtoMajor,
				ProtoMinor: res.ProtoMinor,
				Header:     res.Header,
			}

			hsp.Cache.Put(host, url, response, body, expireTime)
		}
	}

	hsp.CopyHeaders(w.Header(), res.Header)
	w.Header().Del("Date")

	if bCacheCfg.Enabled {
		ttl := hsp.GetTTL(bCacheCfg.TTL, bCacheCfg.TTLUnits)
		ttlStr := strconv.Itoa(ttl)

		w.Header().Set("Cache-Control", "public, max-age="+ttlStr)
	} else {
		w.Header().Del("Cache-Control")
	}

	w.Write(body)
}

func (hsp *httpsProxy) handler(w http.ResponseWriter, r *http.Request, hostCfg config.Domain) {
	tlsCfg := certificate.Generate(r)
	conn, err := hsp.httpsConn(r, hostCfg, tlsCfg)
	if err != nil {
		log.Printf("connection error: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	res, err := hsp.Request(conn, r)
	if err != nil {
		log.Printf("request error: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	hsp.CopyHeaders(w.Header(), res.Header)

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("body error: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(body)
}

func (hsp *httpsProxy) httpsConn(r *http.Request, hostCfg config.Domain,
	cfg *tls.Config) (net.Conn, error) {
	ip := hostCfg.IP
	timeout := time.Duration(hostCfg.Timeout) * time.Second
	dialer := &net.Dialer{
		Timeout: timeout,
	}

	conn, err := tls.DialWithDialer(dialer, "tcp", ip, cfg)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// New creates new https proxy server
func New(domains config.Domains, cache cache.Cacher, port string) *http.Server {
	return &http.Server{
		Addr:         port,
		Handler:      &httpsProxy{handler.Proxy{cache, domains}},
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
	}
}
