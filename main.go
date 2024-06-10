package main

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

var (
	client = &http.Client{}

	backends = map[string]Backend{
		"first": {
			Address: "http://localhost:8090",
		},
	}
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.HandleFunc("/", RootHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

type Client struct {
	http.Client
}

type Backend struct {
	Address     string
	Port        int
	HealthCheck HealthChecks
}
type HealthChecks struct {
	Path               string
	Timeout            int
	ExpectedStatusCode int
}

func RootHandler(w http.ResponseWriter, r *http.Request) {
	u, err := url.Parse(backends["first"].Address)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	r.RequestURI = ""
	r.URL.Host = u.Host
	r.URL.Scheme = u.Scheme

	res, err := client.Do(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	for k, v := range res.Header {
		w.Header()[k] = v
	}
	body := res.Body
	defer body.Close()
	b, err := io.ReadAll(body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(b)
	LogRequest(r)
}

func LogRequest(r *http.Request) {
	log.Println("Received request from", r.RemoteAddr)
	log.Printf("%s %s %s\n", r.Method, r.URL, r.Proto)
	log.Printf("Host: %s\n", r.Host)
	log.Println("User-Agent:", r.UserAgent())
	log.Println("Accept:", r.Header.Get("Accept"))
}
