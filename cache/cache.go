package cache

import "sync"

// Cache is the interface that provides cache methods
type Cache interface {
	Get(key string) []byte
	Put(key string, content []byte)
	Has(key string) bool
}

type cache struct {
	values map[string][]byte
	mutex  *sync.Mutex
}

func (c *cache) Get(key string) []byte {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.values[key]
}

func (c *cache) Put(key string, content []byte) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.values[key] = content
}

func (c *cache) Has(key string) bool {
	_, ok := c.values[key]

	return ok
}

// New creates new instance of cache.
func New() Cache {
	return &cache{
		values: make(map[string][]byte),
		mutex:  &sync.Mutex{},
	}
}
