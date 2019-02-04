package handler

import (
	"bytes"
	"io"
	"log"
	"net/http"

	"github.com/artistomin/proxy/internal/app/proxy/cache"
	"github.com/artistomin/proxy/internal/app/proxy/config"
)

// Handler common structure for handler
type Handler struct {
	Cache  cache.Cacher
	Config *config.Config
	Tr     *http.Transport
}

func (h *Handler) FromCache(w http.ResponseWriter, r *http.Request) {
	url := r.URL.String()
	cachedValue := h.Cache.Get(r.Host, url)

	res := &http.Response{
		Status:     cachedValue.Response.Status,
		StatusCode: cachedValue.Response.StatusCode,
		Proto:      cachedValue.Response.Proto,
		ProtoMajor: cachedValue.Response.ProtoMajor,
		ProtoMinor: cachedValue.Response.ProtoMinor,
		Header:     cachedValue.Response.Header,
	}
	bodyReader := bytes.NewReader(cachedValue.Body)

	h.CopyHeaders(w.Header(), res.Header)

	bytes, err := io.Copy(w, bodyReader)
	if err != nil {
		log.Printf("cache error: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("From cache: %s, bytes: %d", url, bytes)
}

func (h *Handler) DefaultHandler(w http.ResponseWriter, r *http.Request) {
	res, err := h.Tr.RoundTrip(r)
	if err != nil {
		log.Printf("request error: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	h.CopyHeaders(w.Header(), res.Header)

	_, err = io.Copy(w, res.Body)
	if err != nil {
		log.Printf("copy error: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// ShouldResCached checks should be response cached or not
func (h *Handler) ShouldResCached(host, path string, bodySize int, cacheCfg config.Cache) bool {
	if !cacheCfg.Enabled {
		return false
	}

	if len(cacheCfg.Cached) > 0 && !pathHasSuffix(path, cacheCfg.Cached) {
		return false
	}

	if len(cacheCfg.NoCached) > 0 && pathContainsString(path, cacheCfg.NoCached) {
		return false
	}

	if bodySize > cacheCfg.CacheObject.MaxSizeBytes {
		return false
	}

	if (h.Cache.Size(host) + bodySize) >= cacheCfg.MaxSizeBytes {
		return false
	}

	return true
}

// LogRequest logging request
func (h *Handler) LogRequest(r *http.Request, scheme string) {
	log.Printf("Scheme: %s, Method: %s, Host: %s, Url: %s\n", scheme, r.Method, r.Host,
		r.URL.String())
}
