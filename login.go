package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/caleb-fringer/chirpy/internal/auth"
	"github.com/google/uuid"
)

type loginResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Token     string    `json:"token"`
}

type loginRequestParams struct {
	createUserReqParams
	ExpiresInSeconds int `json:"expires_in_seconds,omitempty"`
}

func (cfg *apiConfig) login(w http.ResponseWriter, r *http.Request) {
	reqParams := &loginRequestParams{}
	reqDecoder := json.NewDecoder(r.Body)
	err := reqDecoder.Decode(reqParams)
	if err != nil {
		log.Printf("POST /api/login: Error decoding request: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(json.RawMessage(`"error": "Server error decoding request."`))
		return
	}

	user, err := cfg.queries.GetUserByEmail(r.Context(), reqParams.Email)
	if err != nil {
		log.Printf("POST /api/login: Error retrieving user from database: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(json.RawMessage(`"error": "Database error"`))
		return
	}

	err = auth.CheckPasswordHash(reqParams.Password, user.HashedPassword)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write(json.RawMessage(`"error": "Incorrect email or password"`))
		return
	}

	tokenDuration := time.Second * time.Duration(reqParams.ExpiresInSeconds)
	if tokenDuration == 0 || tokenDuration > time.Hour {
		tokenDuration = time.Hour
	}

	token, err := auth.MakeJWT(user.ID, cfg.secretKey, tokenDuration)
	if err != nil {
		log.Printf("POST /api/login: Error making JWT: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(json.RawMessage(`"error": "Error creating JWT"`))
		return
	}

	response := &loginResponse{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
		Token:     token,
	}

	resJson, err := json.Marshal(response)
	if err != nil {
		log.Printf("POST /api/login: Error encoding response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(json.RawMessage(`"error": "Server error encoding response."`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resJson)
	return
}
