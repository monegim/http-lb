package main

import (
	"log"
	"net/http"
	"os"
)

var client = &http.Client{}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	//backends := map[string]Backend{
	//	"first": {
	//		Address: "http://localhost",
	//		Port:    8090,
	//	},
	//}
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

}
