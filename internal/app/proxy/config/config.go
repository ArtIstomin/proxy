package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"time"
)

const sizeValue = 1024

// Config represents config structure
type Config struct {
	MaxConn     int           `json:"max_connections"`
	Timeout     time.Duration `json:"timeout"`
	IdleTimeout time.Duration `json:"idle_timeout"`
	KeepAlive   time.Duration `json:"keep_alive"`
	Domains     `json:"domains"`
}

// Domains represents domains structure
type Domains map[string]*Domain

// Domain represenets structure of each domain in the config
type Domain struct {
	IP           string `json:"ip"`
	BrowserCache `json:"browser_cache"`
	Cache        `json:"cache"`
}

// Cache represes structure of cache config
type Cache struct {
	Enabled     bool     `json:"enabled"`
	Cached      []string `json:"cached,omitempty"`
	NoCached    []string `json:"no_cached,omitempty"`
	CacheObject size     `json:"cache_object"`
	size
	ttl
}

type BrowserCache struct {
	Enabled bool `json:"enabled"`
	ttl
}

type ttl struct {
	TTLSeconds int           `json:"-"`
	TTL        time.Duration `json:"ttl"`
	TTLUnits   string        `json:"ttl_units"`
}

type size struct {
	MaxSizeBytes int    `json:"-"`
	MaxSize      int    `json:"max_size"`
	SizeUnits    string `json:"size_units"`
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

	for _, domain := range cfg.Domains {
		cache := &domain.Cache

		cache.CacheObject.MaxSizeBytes = maxSizeBytes(cache.CacheObject.MaxSize,
			cache.CacheObject.SizeUnits)
		cache.MaxSizeBytes = maxSizeBytes(cache.MaxSize, cache.SizeUnits)
		cache.TTLSeconds = convertTTL(cache.TTL, cache.TTLUnits)

		bCache := &domain.BrowserCache
		bCache.TTLSeconds = convertTTL(bCache.TTL, bCache.TTLUnits)

	}

	return cfg, nil
}

func convertTTL(ttl time.Duration, units string) int {
	var ttlDuration time.Duration

	switch units {
	case "h":
		ttlDuration = ttl * time.Hour
	case "s":
		ttlDuration = ttl * time.Second
	case "m":
		ttlDuration = ttl * time.Minute
	}

	return int(ttlDuration.Seconds())
}

func maxSizeBytes(size int, units string) int {
	var maxSize int
	units = strings.ToLower(units)

	switch units {
	case "kb":
		maxSize = size * sizeValue
	case "mb":
		maxSize = size * sizeValue * sizeValue
	case "gb":
		maxSize = size * sizeValue * sizeValue * sizeValue
	}

	return maxSize
}
