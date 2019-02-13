package cache

import (
	"sync"
	"time"
	"unsafe"
)

type InmemoryCache struct {
	values
	*sync.Mutex
}

type values map[string]map[string]Value

func (ic *InmemoryCache) Get(host, url string) Value {
	ic.Lock()
	defer ic.Unlock()

	return ic.values[host][url]
}

func (ic *InmemoryCache) Put(host, url string, response Response, body []byte,
	expires time.Time) {
	ic.Lock()
	defer ic.Unlock()

	v := Value{response, body, expires}

	if _, exists := ic.values[host]; !exists {
		ic.values[host] = map[string]Value{url: v}
	} else {
		ic.values[host][url] = v
	}
}

func (ic *InmemoryCache) Has(host, key string) bool {
	obj, ok := ic.values[host][key]
	now := time.Now().UTC()

	if ok && now.After(obj.Expires) {
		return !ok
	}

	return ok
}

func (ic *InmemoryCache) Size(host string) int {
	// TODO: implement another method to get size of host cache
	// unsafe.Sizeof - anti-pattern
	return int(unsafe.Sizeof(ic.values[host]))
}

// NewInmemory creates new instance of cache in memory.
func NewInmemory() *InmemoryCache {
	return &InmemoryCache{
		make(values),
		&sync.Mutex{},
	}
}
