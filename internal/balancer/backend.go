package balancer

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

type Backend struct {
	URL 		 *url.URL
	Alive 		 bool
	mux 		 sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
}

func (b *Backend) SetAlive(alive bool) {
	b.mux.Lock()
	b.Alive = alive 
	b.mux.Unlock()
}

func (b *Backend) IsAlive() bool {
	b.mux.RLock()
	alive := b.Alive
	b.mux.RUnlock()
	return alive
}

type BackendPool struct {
	backends []*Backend
	current  uint64
}

func NewBackendPool(backendURLs []string) *BackendPool {
	var backends []*Backend
	for _, u := range backendURLs {
		backendURL, err := url.Parse(u)
		if err != nil {
			log.Fatal(err)
		}

		proxy := httputil.NewSingleHostReverseProxy(backendURL)
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
			log.Printf("[%s] %s\n", backendURL.Host, e.Error())
			w.WriteHeader(http.StatusBadGateway)
		}

		backends = append(backends, &Backend{
			URL: backendURL,
			Alive: true,
			ReverseProxy: proxy,
		})
	}
	return &BackendPool{backends: backends}
}

func (p *BackendPool) Next() *Backend {
	next := int(p.current % uint64(len(p.backends)))
	p.current++
	return p.backends[next]
}

func (p *BackendPool) GetNextAlive() *Backend {
	next := p.Next()
	for range p.backends {
		if next.IsAlive() {
			return next
		}
		next = p.Next()
	}
	return nil
}

func (p *BackendPool) MarkBackendStatus(backendURL *url.URL, alive bool) {
	for _, backend := range p.backends {
		if backend.URL.String() == backendURL.String() {
			backend.SetAlive(alive)
			return
		}
	}
	log.Printf("Backend %s not found", backendURL)
}