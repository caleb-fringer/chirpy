package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fsHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) reset(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		w.Write(json.RawMessage(`"error": "Access forbidden"`))
		return
	}

	cfg.fsHits.Store(0)
	result, err := cfg.queries.DeleteUsers(r.Context())

	if err != nil {
		log.Printf("Error deleting users: %v\n", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(json.RawMessage(`"error": "Error deleting users from database"`))
		return
	}

	type DeleteResponse struct {
		RowsAffected int64  `json:"rows_affected"`
		Message      string `json:"message"`
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	rows, _ := result.RowsAffected()
	response := DeleteResponse{
		RowsAffected: rows,
		Message:      fmt.Sprintf("Deleted %d users.\n", rows),
	}

	rawRes, err := json.Marshal(response)
	if err != nil {
		log.Printf("POST /admin/reset: Error marshalling response\n")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(json.RawMessage(`"error": "Server error"`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(rawRes)
	return
}

func (cfg *apiConfig) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	metricsPage := fmt.Sprintf(`
    <html>
        <body>
            <h1>Welcome, Chirpy Admin</h1>
            <p>Chirpy has been visited %d times!</p>
        </body>
    </html>`, cfg.fsHits.Load())
	w.Write([]byte(metricsPage))
}
