package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"
)

// Domains represents config structure
type Domains map[string]Domain

// Domain represenets structure of each domain in the config
type Domain struct {
	Pool         `json:"pool"`
	BrowserCache `json:"browser_cache"`
	Cache        `json:"cache"`
}

type Pool struct {
	IP      string `json:"ip"`
	Timeout int    `json:"timeout"`
	MaxConn int    `json:"max_connections"`
	Secure  bool   `json:"secure"`
}

type Cache struct {
	Enabled     bool          `json:"enabled"`
	TTL         time.Duration `json:"ttl,omitempty"`
	TTLUnits    string        `json:"ttl_units,omitempty"`
	MaxSize     int           `json:"max_size,omitempty"`
	SizeUnits   string        `json:"size_units,omitempty"`
	Cached      []string      `json:"cached,omitempty"`
	NoCached    []string      `json:"no_cached,omitempty"`
	CacheObject CacheObject   `json:"cache_object,omitempty"`
}

type CacheObject struct {
	MaxSize   int    `json:"max_size,omitempty"`
	SizeUnits string `json:"size_units,omitempty"`
}

type BrowserCache struct {
	Enabled  bool          `json:"enabled"`
	TTL      time.Duration `json:"ttl,omitempty"`
	TTLUnits string        `json:"ttl_units,omitempty"`
}

// Load returns Domains structure
func Load(path string) (Domains, error) {
	bytes, err := ioutil.ReadFile(path)

	if err != nil {
		return nil, fmt.Errorf("Config error: %s", err)
	}

	domains := make(Domains)
	err = json.Unmarshal(bytes, &domains)

	if err != nil {
		return nil, fmt.Errorf("Config error: %s", err)
	}

	return domains, nil
}
