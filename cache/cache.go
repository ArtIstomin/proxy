package cache

import (
	"net/http"
	"sync"
	"time"
	"unsafe"
)

// Cacher is the interface that provides cache methods
type Cacher interface {
	Get(host, url string) []byte
	Put(host, url string, content []byte)
	Has(host, key string) bool
	Size(host string) int
}

type Cache struct {
	values map[string]map[string]Value
	mutex  *sync.Mutex
}

type Value struct {
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

func (c *Cache) Get(host, url string) Value {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.values[host][url]
}

func (c *Cache) Put(host, url string, response Response, body []byte, expires time.Time) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	value := Value{response, body, expires}

	if _, exists := c.values[host]; !exists {
		c.values[host] = map[string]Value{url: value}
	} else {
		c.values[host][url] = value
	}
}

func (c *Cache) Has(host, key string) bool {
	obj, ok := c.values[host][key]
	now := time.Now().UTC()

	if ok && obj.expires.After(now) {
		return !ok
	}

	return ok
}

func (c *Cache) Size(host string) int {
	// TODO: implement another method to get size of host cache
	// unsafe.Sizeof - anti-pattern
	return int(unsafe.Sizeof(c.values[host]))
}

// New creates new instance of cache.
func New() *Cache {
	return &Cache{
		values: make(map[string]map[string]Value),
		mutex:  &sync.Mutex{},
	}
}
