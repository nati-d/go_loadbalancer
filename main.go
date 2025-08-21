package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

// Backend represents a backend server
type Backend struct {
	URL          *url.URL
	Alive        bool
	mux          sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
}

// Check if backend is alive
func (b *Backend) IsAlive() bool {
	b.mux.RLock()
	defer b.mux.RUnlock()
	return b.Alive
}

// Set backend status
func (b *Backend) SetAlive(alive bool) {
	b.mux.Lock()
	b.Alive = alive
	b.mux.Unlock()
}

// LoadBalancer struct
type LoadBalancer struct {
	backends []*Backend
	current  uint64
	mux      sync.Mutex
}

// NewLoadBalancer creates a load balancer with given backend URLs
func NewLoadBalancer(urls []string) *LoadBalancer {
	backends := make([]*Backend, 0)
	for _, u := range urls {
		parsedURL, err := url.Parse(u)
		if err != nil {
			log.Fatalf("Invalid backend URL: %s", u)
		}
		proxy := httputil.NewSingleHostReverseProxy(parsedURL)
		backends = append(backends, &Backend{
			URL:          parsedURL,
			Alive:        true,
			ReverseProxy: proxy,
		})
	}
	return &LoadBalancer{backends: backends}
}

// Get next alive backend using round-robin
func (lb *LoadBalancer) NextBackend() *Backend {
	lb.mux.Lock()
	defer lb.mux.Unlock()
	total := len(lb.backends)
	for i := 0; i < total; i++ {
		b := lb.backends[lb.current%uint64(total)]
		lb.current++
		if b.IsAlive() {
			return b
		}
	}
	return nil
}

// Health check each backend every interval
func (lb *LoadBalancer) HealthCheck(interval time.Duration) {
	for {
		for _, b := range lb.backends {
			go func(backend *Backend) {
				resp, err := http.Get(backend.URL.String())
				if err != nil || resp.StatusCode >= 400 {
					backend.SetAlive(false)
					log.Printf("Backend %s is DOWN\n", backend.URL)
				} else {
					backend.SetAlive(true)
					log.Printf("Backend %s is UP\n", backend.URL)
				}
			}(b)
		}
		time.Sleep(interval)
	}
}

// ServeHTTP implements the reverse proxy handler
func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	backend := lb.NextBackend()
	if backend == nil {
		http.Error(w, "No backend available", http.StatusServiceUnavailable)
		return
	}

	log.Printf("Forwarding request %s %s to %s\n", r.Method, r.URL.Path, backend.URL)
	backend.ReverseProxy.ServeHTTP(w, r)
}

// Start backend server (for testing)
func startServer(port, message string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, message)
	})
	server := &http.Server{
		Addr:    port,
		Handler: mux,
	}
	log.Printf("%s running on %s\n", message, port)
	log.Fatal(server.ListenAndServe())
}

func main() {
	// Start backend servers in goroutines
	go startServer(":8081", "Response from Server 1")
	go startServer(":8082", "Response from Server 2")
	go startServer(":8083", "Response from Server 3")

	// Give backends time to start
	time.Sleep(time.Second)

	// Load balancer setup
	backendURLs := []string{
		"http://localhost:8081",
		"http://localhost:8082",
		"http://localhost:8083",
	}
	lb := NewLoadBalancer(backendURLs)

	// Start health check
	go lb.HealthCheck(5 * time.Second)

	// Start load balancer server
	log.Println("Load balancer running on :8080")
	log.Fatal(http.ListenAndServe(":8080", lb))
}
