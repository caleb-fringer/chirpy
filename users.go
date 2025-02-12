package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/caleb-fringer/chirpy/internal/auth"
	"github.com/caleb-fringer/chirpy/internal/database"
	"github.com/google/uuid"
)

type createUserReqParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type createUserResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func validPassword(password string) bool {
	return len(password) > 7
}

func (cfg *apiConfig) createUser(w http.ResponseWriter, r *http.Request) {
	params := &createUserReqParams{}
	rawReqBody, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("POST /api/users: Error decoding requested user data: %v\n", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write(json.RawMessage(`{"error": "Bad email/password"}`))
		return
	}

	json.Unmarshal(rawReqBody, params)

	if !validPassword(params.Password) {
		log.Printf("POST /api/users: Error hashing password: %v\n", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write(json.RawMessage(`{"error": "Invalid password. Minimum password length is 8 characters."}`))
		return
	}

	hash, err := auth.HashPassword(params.Password)
	if err != nil {
		log.Printf("POST /api/users: Error hashing password: %v\n", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write(json.RawMessage(`{"error": "Bad email/password"}`))
		return
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

	res := createUserResponse{
		user.ID,
		user.CreatedAt,
		user.UpdatedAt,
		user.Email,
	}

	rawResBody, err := json.Marshal(res)
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
