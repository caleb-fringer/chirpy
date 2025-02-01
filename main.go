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

	apiCfg := &apiConfig{}

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(ROOT_PATH)))))

	mux.HandleFunc("GET /api/healthz", healthz)

	mux.HandleFunc("POST /admin/reset", apiCfg.reset)

	mux.Handle("GET /admin/metrics", apiCfg)

	mux.HandleFunc("POST /api/validate_chirp", validateChirp)

	fmt.Printf("Starting server on port %d...", PORT)
	log.Fatal(server.ListenAndServe())
}
