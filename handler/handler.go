package handler

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/artistomin/proxy/cache"
)

var (
	http200    = []byte("HTTP/1.1 200 Connection Established\r\n\r\n")
	extensions = [...]string{".jpeg", ".jpg", ".js", ".png"}
)

type proxy struct {
	client *http.Client
	cache  cache.Cache
}

func (p *proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}()

	if r.Method == http.MethodConnect {
		p.handlerHTTPS(w, r)
	} else if r.Method == http.MethodGet && filterAssets(r.URL.String()) {
		p.handlerCache(w, r)
	} else {
		p.handlerHTTP(w, r)
	}
}

func (p *proxy) handlerCache(w http.ResponseWriter, r *http.Request) {

	uri := r.RequestURI
	cached := p.cache.Has(uri)

	if cached {
		logRequest(r, "http")

		content := p.cache.Get(uri)
		w.Write(content)

		log.Printf("Bytes: %d, Host: %s, URL: %s, Cached: %t", len(content), r.URL.Host, r.URL.String(), cached)
		return
	}

	res, err := p.client.Get(uri)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	w.WriteHeader(res.StatusCode)

	body, err := ioutil.ReadAll(res.Body)

	p.cache.Put(uri, body)
	w.Write(body)
}

func (p *proxy) handlerHTTP(w http.ResponseWriter, r *http.Request) {
	rmProxyHeaders(r)
	res, err := p.client.Do(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	copyHeaders(w.Header(), res.Header)
	w.WriteHeader(res.StatusCode)

	_, err = io.Copy(w, res.Body)
	if err != nil && err != io.EOF {
		log.Printf("occur an error when copying remote response to this client, %v", err)
		return
	}
}

func (p *proxy) handlerHTTPS(w http.ResponseWriter, r *http.Request) {
	config := generateCfg(r)

	hj, _ := w.(http.Hijacker)
	hjClient, _, err := hj.Hijack()

	if err != nil {
		log.Printf("http user failed to get tcp connection")
		http.Error(w, "Failed", http.StatusBadRequest)
		return
	}

	hjClient.Write(http200)

	go func() {
		defer hjClient.Close()
		tlsCon := tls.Server(hjClient, config)
		clientTLSReader := bufio.NewReader(tlsCon)
		clientTLSWriter := bufio.NewWriter(tlsCon)

		for !isEOF(clientTLSReader) {
			req, err := http.ReadRequest(clientTLSReader)

			if err != nil && err != io.EOF {
				log.Printf("Read Request Error: %+#v", err.Error())
				return
			}

			if err != nil {
				log.Printf("cannot read request of MITM HTTP client: %+#v", err.Error())
				return
			}

			req.RequestURI = ""
			req.URL = buildHTTPSUrl(req)

			cached := p.cache.Has(req.URL.String())
			var res *http.Response

			if cached && req.Method == http.MethodGet {
				logRequest(req, "https")
				cachedRes := p.cache.Get(req.URL.String())
				bytesReader := bytes.NewReader(cachedRes)
				bufioReader := bufio.NewReader(bytesReader)
				res, err = http.ReadResponse(bufioReader, req)

				if err != nil {
					log.Printf("Read Response error: %v\n", err.Error())
					return
				}

				log.Printf("Bytes: %d, Host: %s, URL: %s, Cached: %t", len(cachedRes), req.URL.Host, req.URL.String(), cached)
			} else {
				resp, err := p.client.Transport.RoundTrip(req)
				if err != nil {
					log.Printf("HTTPS client error: %v\n", err.Error())
					return
				}

				res = resp

				if filterAssets(req.URL.String()) && req.Method == http.MethodGet {
					dumpRes, err := httputil.DumpResponse(res, true)

					if err != nil {
						log.Printf("Dump Response error: %v\n", err.Error())
						return
					}

					p.cache.Put(req.URL.String(), dumpRes)
				}

			}

			err = res.Write(clientTLSWriter)
			if err != nil {
				log.Printf("RES write error: %v\n", err.Error())
				return
			}

			err = clientTLSWriter.Flush()
			if err != nil {
				log.Printf("FLUSH error: %v\n", err.Error())
				return
			}
		}
	}()
}

func buildHTTPSUrl(r *http.Request) *url.URL {
	url, err := url.Parse("https://" + r.Host + r.URL.String())
	if err != nil {
		log.Printf("Build URL error: %v\n", err.Error())
	}
	return url
}

func logRequest(r *http.Request, proto string) {
	log.Printf("Method: %s, Proto: %s, Host:%s, Url: %s\n", r.Method, proto, r.Host, r.URL.String())
}

func isEOF(r *bufio.Reader) bool {
	_, err := r.Peek(1)
	if err == io.EOF {
		return true
	}
	return false
}

func filterAssets(URL string) bool {
	for _, extension := range extensions {
		if strings.Contains(URL, extension) {
			return true
		}
	}

	return false
}

// New creates new proxy server
func New(client *http.Client, cache cache.Cache, port string) *http.Server {
	return &http.Server{
		Addr:    port,
		Handler: &proxy{client, cache},
	}
}
