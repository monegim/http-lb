package main

import (
	"http-lb/internal"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"slices"
	"sync"
	"sync/atomic"
	"time"
)

var (
	serverPool = ServerPool{
		backends: []*Backend{
			{
				Address: "http://localhost:8090",
			}, {
				Address: "http://localhost:8091",
			},
		},
	}
	requestCounter int
	client         = internal.NewRequester()
)

const (
	interval = 5
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.HandleFunc("/", lb)

	go serverPool.healthcheck()
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

type ServerPool struct {
	backends []*Backend
	current  uint64
}

// AddBackend to the server pool
func (s *ServerPool) AddBackend(backend *Backend) {
	s.backends = append(s.backends, backend)
}

func (s *ServerPool) NextIndex() int {
	return int(atomic.AddUint64(&s.current, uint64(1)) % uint64(len(s.backends)))
}

func (s *ServerPool) MarkBackendStatus(backendUrl *url.URL, alive bool) {
	for _, b := range s.backends {
		if b.Address == backendUrl.String() {
			b.SetAlive(alive)
			break
		}
	}
}

func (s *ServerPool) HealthCheckBackends() {
	for _, b := range s.backends {
		status := "up"
		alive := b.isBackendAlive()
		b.SetAlive(alive)
		if !alive {
			status = "down"
		}
		log.Printf("%s [%s]\n", b.Address, status)
	}
}

func (s *ServerPool) GetNextPeer() *Backend {
	nextIndex := s.NextIndex()
	l := len(s.backends) + nextIndex
	for i := nextIndex; i < l; i++ {
		idx := i % len(s.backends)
		b := s.backends[idx]
		if b.isBackendAlive() {
			return b
		}
	}
	return nil
}

func (b *Backend) SetAlive(alive bool) {
	b.mux.RLock()
	b.Alive = alive
	defer b.mux.RUnlock()
	return
}

type Backend struct {
	Address      string
	mux          *sync.RWMutex
	HealthCheck  HealthCheck
	Alive        bool
	ReverseProxy *httputil.ReverseProxy
}
type HealthCheck struct {
	Path               string
	Timeout            int
	ExpectedStatusCode []int
}

func (s *ServerPool) healthcheck() {
	t := time.NewTicker(time.Second * interval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			log.Println("Starting healthcheck...")
			s.HealthCheckBackends()
			log.Println("Health check done.")
		}
	}
}

func lb(w http.ResponseWriter, r *http.Request) {
	peer := serverPool.GetNextPeer()
	if peer == nil {
		http.Error(w, "no peer available", http.StatusServiceUnavailable)
	}
	peer.ReverseProxy.ServeHTTP(w, r)
	//requestCounter++
	//u, err := url.Parse()
	//if err != nil {
	//	w.WriteHeader(http.StatusBadRequest)
	//	w.Write([]byte(err.Error()))
	//	return
	//}
	//
	//r.RequestURI = ""
	//r.URL.Host = u.Host
	//r.URL.Scheme = u.Scheme
	//
	//res, err := client.HttpClient.Do(r)
	//if err != nil {
	//	w.WriteHeader(http.StatusInternalServerError)
	//	w.Write([]byte(err.Error()))
	//	return
	//}
	//for k, v := range res.Header {
	//	w.Header()[k] = v
	//}
	//body := res.Body
	//defer body.Close()
	//b, err := io.ReadAll(body)
	//if err != nil {
	//	w.WriteHeader(http.StatusInternalServerError)
	//	w.Write([]byte(err.Error()))
	//	return
	//}
	//w.Write(b)
	//LogRequest(r)
}

func LogRequest(r *http.Request) {
	log.Println("Received request from", r.RemoteAddr)
	log.Printf("%s %s %s\n", r.Method, r.URL, r.Proto)
	log.Printf("Host: %s\n", r.Host)
	log.Println("User-Agent:", r.UserAgent())
	log.Println("Accept:", r.Header.Get("Accept"))
}

func (b *Backend) isBackendAlive() bool {
	endpoint := b.Address + b.HealthCheck.Path
	res, err := client.HttpClient.Get(endpoint)
	if err != nil {
		return false
	}
	if slices.Contains(b.HealthCheck.ExpectedStatusCode, res.StatusCode) {
		return true
	}
	return false
}
