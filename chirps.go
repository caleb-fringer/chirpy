package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/caleb-fringer/chirpy/internal/database"
)

func (cfg *apiConfig) createChirp(w http.ResponseWriter, r *http.Request) {
	reqBody := &database.CreateChirpParams{}

	err := json.NewDecoder(r.Body).Decode(reqBody)
	if err != nil {
		log.Printf("POST /api/chirps: Error decoding request body: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(json.RawMessage(`"error": "Server error decoding request body"`))
		return
	}

	censored, ok := validateChirp(reqBody.Body)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(json.RawMessage(`"error": "Chirp is too long. Max chirp length is 140 characters."`))
		return
	}
	reqBody.Body = censored

	chirp, err := cfg.queries.CreateChirp(r.Context(), *reqBody)

	if err != nil {
		log.Printf("POST /api/chirps: Error creating chirp: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(json.RawMessage(`"error": "Database error"`))
		return
	}

	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(chirp)
	if err != nil {
		log.Printf("POST /api/chirps: Error writing chirp response: %v\n", err)
	}
	return
}

func (cfg *apiConfig) getChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.queries.GetChirps(r.Context())
	if err != nil {
		log.Printf("GET /api/chirps: Error retrieving chirps: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(json.RawMessage(`"error": "Database error"`))
		return
	}

	jsonRes, err := json.Marshal(chirps)
	if err != nil {
		log.Printf("GET /api/chirps: Error encoding chirps response: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(json.RawMessage(`"error": "Encoding error"`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonRes)
	return
}
