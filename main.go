package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/caleb-fringer/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const PORT = 8080
const ROOT_PATH = "."

type apiConfig struct {
	platform  string
	fsHits    atomic.Int32
	queries   *database.Queries
	secretKey string
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	secretKey := os.Getenv("SECRET_KEY")
	db, err := sql.Open("postgres", dbURL)

	if err != nil {
		log.Fatalf("Error connecting to the database: %v\n", err)
	}

	dbQueries := database.New(db)
	apiCfg := &apiConfig{
		platform:  platform,
		queries:   dbQueries,
		secretKey: secretKey,
	}

	mux := http.NewServeMux()
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", PORT),
		Handler: mux,
	}

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(ROOT_PATH)))))

	mux.HandleFunc("GET /api/healthz", healthz)

	mux.HandleFunc("POST /api/users", apiCfg.createUser)

	mux.HandleFunc("POST /admin/reset", apiCfg.reset)

	mux.Handle("GET /admin/metrics", apiCfg)

	mux.HandleFunc("POST /api/chirps", apiCfg.createChirp)

	mux.HandleFunc("GET /api/chirps", apiCfg.getChirps)

	mux.HandleFunc("GET /api/chirps/{id}", apiCfg.getChirp)

	mux.HandleFunc("POST /api/login", apiCfg.login)

	mux.HandleFunc("POST /api/refresh", apiCfg.refresh)

	mux.HandleFunc("POST /api/revoke", apiCfg.handlerRevoke)

	fmt.Printf("Starting server on port %d...\n", PORT)
	log.Fatal(server.ListenAndServe())
}
