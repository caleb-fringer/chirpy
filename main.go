package main

import (
	"fmt"
	"log"
	"net/http"
)

const PORT = 8080
const ROOT_PATH = "."

func main() {
	mux := http.NewServeMux()
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", PORT),
		Handler: mux,
	}

	mux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir(ROOT_PATH))))

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	fmt.Printf("Starting server on port %d...", PORT)
	log.Fatal(server.ListenAndServe())
}
