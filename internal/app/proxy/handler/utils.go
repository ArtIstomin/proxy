package handler

import (
	"log"
	"net/http"
	"strings"
	"time"
)

// GetTTL converts ttl to seconds
func (p *Proxy) GetTTL(ttl time.Duration, units string) int {
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

// LogRequest logging request
func (p *Proxy) LogRequest(r *http.Request, scheme string) {
	log.Printf("Scheme: %s, Method: %s, Host: %s, Url: %s\n", scheme, r.Method, r.Host,
		r.URL.String())
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

func pathHasSuffix(path string, suffixes []string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(path, suffix) {
			return true
		}
	}

	return false
}

func pathContainsString(path string, subStrings []string) bool {
	for _, subString := range subStrings {
		if strings.Contains(path, subString) {
			return true
		}
	}

	return false
}
