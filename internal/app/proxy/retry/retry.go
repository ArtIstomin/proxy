package retry

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/artistomin/proxy/internal/app/proxy/cache"
	"github.com/artistomin/proxy/internal/app/proxy/config"
	pb "github.com/artistomin/proxy/internal/pkg/proto/activity"
	"github.com/golang/protobuf/ptypes/empty"
)

type Retry struct {
	Cache  cache.Cacher
	Tr     *http.Transport
	Config *config.Config
	client pb.ActivityClient
}

type request struct {
}

func (r *Retry) ProcessUncompletedReqs() {
	resp, err := r.client.GetRequests(context.Background(), &empty.Empty{})

	if err != nil {
		log.Printf("retrier error: %s", err)
		return
	}

	for _, req := range resp.Requests {
		go r.processReq(req)
	}
}

func (r *Retry) processReq(pbreq *pb.ReqRequest) {
	req, err := http.NewRequest("GET", pbreq.Url, nil)
	if err != nil {
		log.Printf("retrier error: new request: %s", err)
		return
	}

	if err := json.Unmarshal(pbreq.Header, &req.Header); err != nil {
		log.Printf("retrier error: unmarshal headers: %s", err)
		return
	}

	res, err := r.Tr.RoundTrip(req)
	if err != nil {
		log.Printf("retrier error: do request: %s", err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("retrier error: body error: %s", err)
		return
	}

	host := req.Host
	cacheCfg := r.Config.Domains[req.Host].Cache
	url := req.URL.String()
	if r.shouldResCached(host, req.URL.Path, len(body), cacheCfg) {
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

		r.Cache.Put(host, url, response, body, expireTime)
	}

	payloadUpdate := &pb.ReqRequest{
		ReqId:     int32(pbreq.ReqId),
		Completed: true,
	}
	_, err = r.client.UpdateRequest(context.Background(), payloadUpdate)
	if err != nil {
		return
	}
}

func (r *Retry) shouldResCached(host, path string, bodySize int, cacheCfg config.Cache) bool {
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

	if (r.Cache.Size(host) + bodySize) >= cacheCfg.MaxSizeBytes {
		return false
	}

	return true
}

// New returns new retrier
func New(cache cache.Cacher, tr *http.Transport, cfg *config.Config, client pb.ActivityClient) *Retry {
	return &Retry{cache, tr, cfg, client}
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
