package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

var Version = "dev"

func main() {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok","version":"%s"}`, Version)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"message":"Trello Agent API","version":"%s"}`, Version)
	})

	log.Printf("Trello Agent API starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
