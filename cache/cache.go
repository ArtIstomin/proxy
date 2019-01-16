package cache

import "sync"

// Cacher is the interface that provides cache methods
type Cacher interface {
	Get(host, url string) []byte
	Put(host, url string, content []byte)
	Has(host, key string) bool
}

type Cache struct {
	values map[string]map[string][]byte
	mutex  *sync.Mutex
}

func (c *Cache) Get(host, url string) []byte {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.values[host][url]
}

func (c *Cache) Put(host, url string, content []byte) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.values[host][url] = content
}

func (c *Cache) Has(host, key string) bool {
	_, ok := c.values[host][key]

	return ok
}

func (c *Cache) Size(host string) int {
	return len(c.values[host])
}

// New creates new instance of cache.
func New() *Cache {
	return &Cache{
		values: make(map[string]map[string][]byte),
		mutex:  &sync.Mutex{},
	}
}
