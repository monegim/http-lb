package main

import (
	"http-lb/handlers"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.HandleFunc("/", handlers.RootHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
