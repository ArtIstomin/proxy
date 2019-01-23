package handler

import (
	"bufio"
	"net"
	"net/http"
	"net/http/httputil"

	"github.com/artistomin/proxy/internal/app/proxy/cache"
	"github.com/artistomin/proxy/internal/app/proxy/config"
)

const sizeValue = 1024

// Handler common structure for handler
type Handler struct {
	Cache   cache.Cacher
	Domains config.Domains
}

// Request performs request to destination server
func (h *Handler) Request(conn net.Conn, r *http.Request) (*http.Response, error) {
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

// ShouldResCached checks should be response cached or not
func (h *Handler) ShouldResCached(host, path string, bodySize int, cacheCfg config.Cache) bool {
	if !cacheCfg.Enabled {
		return false
	}

	if (h.Cache.Size(host) + bodySize) >= maxSizeBytes(cacheCfg.MaxSize, cacheCfg.SizeUnits) {
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
