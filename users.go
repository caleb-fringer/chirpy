package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
)

type reqParams struct {
	Email string `json:"email"`
}

type createdUser struct {
	Id        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) createUser(w http.ResponseWriter, r *http.Request) {
	params := &reqParams{}
	rawReqBody, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("POST /api/users: Error decoding requested user email: %v\n", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write(json.RawMessage(`{"error": "Bad email"}`))
		return
	}

	json.Unmarshal(rawReqBody, params)
	user, err := cfg.queries.CreateUser(r.Context(), params.Email)

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
