package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"
)

// Config represents config structure
type Config struct {
	MaxConn     int `json:"max_connections"`
	Timeout     int `json:"timeout"`
	IdleTimeout int `json:"idle_timeout"`
	KeepAlive   int `json:"keep_alive"`
	Domains     `json:"domains"`
}

// Domains represents domains structure
type Domains map[string]Domain

// Domain represenets structure of each domain in the config
type Domain struct {
	IP           string `json:"ip"`
	BrowserCache `json:"browser_cache"`
	Cache        `json:"cache"`
}

// Cache represes structure of cache config
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
func Load(path string) (*Config, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Config error: %s", err)
	}

	cfg := new(Config)
	err = json.Unmarshal(bytes, &cfg)
	if err != nil {
		return nil, fmt.Errorf("Config error: %s", err)
	}

	return cfg, nil
}
