package handlerhttps

import (
	"context"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/artistomin/proxy/internal/app/proxy/cache"
	"github.com/artistomin/proxy/internal/app/proxy/certificate"
	"github.com/artistomin/proxy/internal/app/proxy/config"
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

	host := r.Host
	hostCfg := hsh.Config.Domains[host]
	dialer := &net.Dialer{
		Timeout:   hsh.Config.Timeout * time.Second,
		KeepAlive: hsh.Config.KeepAlive * time.Second,
	}
	hsh.Tr.TLSClientConfig = certificate.Generate(host)
	hsh.Tr.DialContext = func(ctx context.Context, network, ip string) (net.Conn, error) {
		ip = hostCfg.IP
		return dialer.DialContext(ctx, network, ip)
	}

	r.URL, _ = url.Parse("https://" + host + r.URL.String())

	hsh.RmProxyHeaders(r)
	hsh.LogRequest(r, "https")

	if r.Method == http.MethodGet {
		hsh.handlerCache(w, r, hostCfg)
	} else {
		hsh.DefaultHandler(w, r)
	}
}

func (hsh *HttpsHandler) handlerCache(w http.ResponseWriter, r *http.Request,
	hostCfg *config.Domain) {
	host := r.Host
	url := r.URL.String()
	cacheCfg := hostCfg.Cache
	bCacheCfg := hostCfg.BrowserCache

	if bCacheCfg.Enabled {
		ttlStr := strconv.Itoa(bCacheCfg.TTLSeconds)

		w.Header().Set("Cache-Control", "public, max-age="+ttlStr)
	} else {
		w.Header().Del("Cache-Control")
	}

	switch {
	case cacheCfg.Enabled && hsh.Cache.Has(host, url):
		hsh.FromCache(w, r)
	default:
		res, err := hsh.Tr.RoundTrip(r)
		if err != nil {
			log.Printf("request error: %s", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer res.Body.Close()

		bodyReader := io.TeeReader(res.Body, w)

		hsh.CopyHeaders(w.Header(), res.Header)

		body, err := ioutil.ReadAll(bodyReader)
		if err != nil {
			log.Printf("body error: %s", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if hsh.ShouldResCached(host, r.URL.Path, len(body), cacheCfg) {
			ttl := time.Duration(cacheCfg.TTLSeconds) * time.Second
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
}

// New creates new https handler
func New(cfg *config.Config, cache cache.Cacher, tr *http.Transport) *HttpsHandler {
	return &HttpsHandler{handler.Handler{cache, cfg, tr}}
}
