package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/caleb-fringer/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const PORT = 8080
const ROOT_PATH = "."

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)

	if err != nil {
		log.Fatalf("Error connecting to the database: %v\n", err)
	}

	dbQueries := database.New(db)
	apiCfg := &apiConfig{
		queries: dbQueries,
	}

	mux := http.NewServeMux()
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", PORT),
		Handler: mux,
	}

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(ROOT_PATH)))))

	mux.HandleFunc("GET /api/healthz", healthz)

	mux.HandleFunc("POST /admin/reset", apiCfg.reset)

	mux.Handle("GET /admin/metrics", apiCfg)

	mux.HandleFunc("POST /api/validate_chirp", validateChirp)

	fmt.Printf("Starting server on port %d...", PORT)
	log.Fatal(server.ListenAndServe())
}
