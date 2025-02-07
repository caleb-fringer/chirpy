package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/caleb-fringer/chirpy/internal/auth"
	"github.com/caleb-fringer/chirpy/internal/database"
)

type reqParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (cfg *apiConfig) createUser(w http.ResponseWriter, r *http.Request) {
	params := &reqParams{}
	rawReqBody, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("POST /api/users: Error decoding requested user data: %v\n", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write(json.RawMessage(`{"error": "Bad email/password"}`))
		return
	}

	json.Unmarshal(rawReqBody, params)
	hash, err := auth.HashPassword(params.Password)
	if err != nil {
		log.Printf("POST /api/users: Error hashing password: %v\n", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write(json.RawMessage(`{"error": "Bad email/password"}`))
	}

	userParams := database.CreateUserParams{Email: params.Email, HashedPassword: hash}

	user, err := cfg.queries.CreateUser(r.Context(), userParams)

	if err != nil {
		log.Printf("POST /api/users: Error creating user %s in database: %v\n", params.Email, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write(json.RawMessage(`{"error": "Database error"}`))
		return
	}

	rawResBody, err := json.Marshal(user)
	if err != nil {
		log.Printf("POST /api/users: Error encoding user %s to binary for response: %v\n", params.Email, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write(json.RawMessage(`{"error": "Encoding error"}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(rawResBody)
	return
}
