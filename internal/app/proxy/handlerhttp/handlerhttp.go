package handlerhttp

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
	"github.com/artistomin/proxy/internal/app/proxy/config"
	"github.com/artistomin/proxy/internal/app/proxy/handler"
	pb "github.com/artistomin/proxy/internal/pkg/proto/activity"
)

type HttpHandler struct {
	handler.Handler
}

func (hh *HttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("panic: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}()

	host := r.Host
	hostCfg := hh.Config.Domains[host]
	dialer := &net.Dialer{
		Timeout:   hh.Config.Timeout * time.Second,
		KeepAlive: hh.Config.KeepAlive * time.Second,
	}
	hh.Tr.DialContext = func(ctx context.Context, network, ip string) (net.Conn, error) {
		ip = hostCfg.IP
		return dialer.DialContext(ctx, network, ip)
	}

	r.URL, _ = url.Parse("http://" + host + r.URL.String())

	hh.RmProxyHeaders(r)
	hh.LogRequest(r, "http")

	if r.Method == http.MethodGet {
		hh.handlerCache(w, r, hostCfg)
	} else {
		hh.DefaultHandler(w, r)
	}
}

func (hh *HttpHandler) handlerCache(w http.ResponseWriter, r *http.Request,
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
	case cacheCfg.Enabled && hh.Cache.Has(host, url):
		hh.FromCache(w, r)
	default:
		reqID, err := hh.StoreRequest(r)
		if err != nil {
			log.Printf("store request error: %s", err)
		}

		res, err := hh.Tr.RoundTrip(r)
		if err != nil {
			log.Printf("request error: %s", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer res.Body.Close()

		err = hh.MarkReqCompleted(reqID)
		if err != nil {
			log.Printf("update request error: %s", err)
		}

		bodyReader := io.TeeReader(res.Body, w)

		hh.CopyHeaders(w.Header(), res.Header)

		body, err := ioutil.ReadAll(bodyReader)
		if err != nil {
			log.Printf("body error: %s", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if hh.ShouldResCached(host, r.URL.Path, len(body), cacheCfg) {
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

			hh.Cache.Put(host, url, response, body, expireTime)
		}
	}
}

// New creates new http handler
func New(cfg *config.Config, cache cache.Cacher, tr *http.Transport, client pb.ActivityClient) *HttpHandler {
	return &HttpHandler{handler.Handler{cache, cfg, tr, client}}
}
