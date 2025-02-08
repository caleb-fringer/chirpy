package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/caleb-fringer/chirpy/internal/auth"
	"github.com/caleb-fringer/chirpy/internal/database"
	"github.com/google/uuid"
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

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("POST /api/chirps: Error retrieving bearer token from authorization header: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		switch err.(type) {
		case auth.HeaderNotFoundError:
			w.Write(json.RawMessage(`"error": "Please provide your JWT in the authorization header of your request."`))
		case auth.WrongAuthorizationSchemeError:
			w.Write(json.RawMessage(`"error": "Please use the Bearer authorization scheme to authorize your request."`))
		}
		return
	}

	id, err := auth.ValidateJWT(token, cfg.secretKey)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write(json.RawMessage(`"error": "Invalid token."`))
		return
	}
	reqBody.UserID.UUID = id
	reqBody.UserID.Valid = true

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
		w.Write(json.RawMessage(`"error": "Error encoding response"`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonRes)
	return
}

func (cfg *apiConfig) getChirp(w http.ResponseWriter, r *http.Request) {
	log.Printf("UUID: %v\n", r.PathValue("id"))
	id := uuid.MustParse(r.PathValue("id"))
	chirp, err := cfg.queries.GetChirp(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write(json.RawMessage(`"error": "Chirp not found"`))
		return
	}

	jsonRes, err := json.Marshal(chirp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(json.RawMessage(`"error": "Error encoding response"`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonRes)
	return
}
