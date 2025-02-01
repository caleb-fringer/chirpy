package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

const PORT = 8080
const ROOT_PATH = "."

type apiConfig struct {
	fsHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fsHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) reset(w http.ResponseWriter, r *http.Request) {
	cfg.fsHits.Store(0)
}

func (cfg *apiConfig) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Hits: %d\n", cfg.fsHits.Load())))
}

func main() {
	mux := http.NewServeMux()
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", PORT),
		Handler: mux,
	}

	apiCfg := &apiConfig{}

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(ROOT_PATH)))))

	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("POST /api/reset", apiCfg.reset)

	mux.Handle("GET /api/metrics", apiCfg)

	fmt.Printf("Starting server on port %d...", PORT)
	log.Fatal(server.ListenAndServe())
}
