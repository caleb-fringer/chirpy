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
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Email       string    `json:"email"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
}

func validPassword(password string) bool {
	return len(password) > 0
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
		user.IsChirpyRed,
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

func (cfg *apiConfig) updateUser(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("PUT /api/users: Error reading request headers: %v\n", err)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write(json.RawMessage(`{"error": "Missing/malformed access token."}`))
		return
	}

	uuid, err := auth.ValidateJWT(token, cfg.secretKey)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write(json.RawMessage(`{"error": "Missing/malformed access token."}`))
		return
	}

	rawReqBody, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("PUT /api/users: Error reading request body: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(json.RawMessage(`{"error": "Server error"}`))
		return
	}

	reqBody := &struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}

	err = json.Unmarshal(rawReqBody, reqBody)
	if err != nil {
		log.Printf("PUT /api/users: Error unmarshalling request body: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(json.RawMessage(`{"error": "Server error"}`))
		return
	}

	hashedPassword, err := auth.HashPassword(reqBody.Password)
	if err != nil {
		log.Printf("PUT /api/users: Error hashing password: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(json.RawMessage(`{"error": "Server error"}`))
		return
	}

	updateUserParams := database.UpdateUsernamePasswordParams{
		ID:             uuid,
		Email:          reqBody.Email,
		HashedPassword: hashedPassword,
	}
	user, err := cfg.queries.UpdateUsernamePassword(r.Context(), updateUserParams)
	if err != nil {
		log.Printf("PUT /api/users: Error updating user %d: %v\n", uuid, err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(json.RawMessage(`{"error": "Server error"}`))
		return
	}

	rawRes, err := json.Marshal(&user)
	if err != nil {
		log.Printf("PUT /api/users: Error marshalling response: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(json.RawMessage(`{"error": "Server error"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(rawRes)
	return
}
