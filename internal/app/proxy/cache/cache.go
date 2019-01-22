package cache

import (
	"net/http"
	"sync"
	"time"
	"unsafe"
)

// Cacher is the interface that provides cache methods
type Cacher interface {
	Get(host, url string) value
	Put(host, url string, response Response, body []byte, expires time.Time)
	Has(host, key string) bool
	Size(host string) int
}

type cache struct {
	values
	*sync.Mutex
}

type values map[string]map[string]value

type value struct {
	Response Response
	Body     []byte
	expires  time.Time
}

type Response struct {
	Status     string
	StatusCode int
	Proto      string
	ProtoMajor int
	ProtoMinor int
	Header     http.Header
}

func (c *cache) Get(host, url string) value {
	c.Lock()
	defer c.Unlock()

	return c.values[host][url]
}

func (c *cache) Put(host, url string, response Response, body []byte, expires time.Time) {
	c.Lock()
	defer c.Unlock()

	v := value{response, body, expires}

	if _, exists := c.values[host]; !exists {
		c.values[host] = map[string]value{url: v}
	} else {
		c.values[host][url] = v
	}
}

func (c *cache) Has(host, key string) bool {
	obj, ok := c.values[host][key]
	now := time.Now().UTC()

	if ok && obj.expires.After(now) {
		return !ok
	}

	return ok
}

func (c *cache) Size(host string) int {
	// TODO: implement another method to get size of host cache
	// unsafe.Sizeof - anti-pattern
	return int(unsafe.Sizeof(c.values[host]))
}

// New creates new instance of cache.
func New() Cacher {
	return &cache{
		make(map[string]map[string]value),
		&sync.Mutex{},
	}
}
