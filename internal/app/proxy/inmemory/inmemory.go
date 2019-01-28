package inmemory

import (
	"sync"
	"time"
	"unsafe"

	"github.com/artistomin/proxy/internal/app/proxy/cache"
)

type inmemoryCache struct {
	values
	*sync.Mutex
}

type values map[string]map[string]cache.Value

func (ic *inmemoryCache) Get(host, url string) cache.Value {
	ic.Lock()
	defer ic.Unlock()

	return ic.values[host][url]
}

func (ic *inmemoryCache) Put(host, url string, response cache.Response, body []byte,
	expires time.Time) {
	ic.Lock()
	defer ic.Unlock()

	v := cache.Value{response, body, expires}

	if _, exists := ic.values[host]; !exists {
		ic.values[host] = map[string]cache.Value{url: v}
	} else {
		ic.values[host][url] = v
	}
}

func (ic *inmemoryCache) Has(host, key string) bool {
	obj, ok := ic.values[host][key]
	now := time.Now().UTC()

	if ok && obj.Expires.After(now) {
		return !ok
	}

	return ok
}

func (ic *inmemoryCache) Size(host string) int {
	// TODO: implement another method to get size of host cache
	// unsafe.Sizeof - anti-pattern
	return int(unsafe.Sizeof(ic.values[host]))
}

// New creates new instance of cache.
func New() *inmemoryCache {
	return &inmemoryCache{
		make(values),
		&sync.Mutex{},
	}
}
