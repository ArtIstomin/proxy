package cache

import (
	"net/http"
	"time"
)

// Cacher is the interface that provides cache methods
type Cacher interface {
	Get(host, url string) Value
	Put(host, url string, response Response, body []byte, expires time.Time)
	Has(host, key string) bool
	Size(host string) int
}

type Value struct {
	Response Response
	Body     []byte
	Expires  time.Time
}

type Response struct {
	Status     string      `json:"status"`
	StatusCode int         `json:"status_code"`
	Proto      string      `json:"proto"`
	ProtoMajor int         `json:"proto_major"`
	ProtoMinor int         `json:"proto_minor`
	Header     http.Header `json:"header"`
}
