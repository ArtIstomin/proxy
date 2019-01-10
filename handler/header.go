package handler

import "net/http"

func copyHeaders(dst, src http.Header) {
	for key, values := range src {
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

func rmProxyHeaders(r *http.Request) {
	r.RequestURI = ""
}
