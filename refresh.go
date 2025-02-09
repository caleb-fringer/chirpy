package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/caleb-fringer/chirpy/internal/auth"
	"github.com/caleb-fringer/chirpy/internal/database"
)

func (cfg *apiConfig) refresh(w http.ResponseWriter, r *http.Request) {
	tokenStr, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("POST /api/refresh: Error extracting refresh token from request headers: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(json.RawMessage(`"error": "Please provide a refresh token in the authorization header."`))
		return
	}

	token, err := cfg.queries.GetRefreshTokenById(r.Context(), tokenStr)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write(json.RawMessage(`"error": "Invalid token."`))
		return
	}

	if time.Now().UTC().After(token.ExpiresAt) {
		fmt.Println("Now:", time.Now().UTC())
		fmt.Println("Token ExpiresAt:", token.ExpiresAt)

		w.WriteHeader(http.StatusUnauthorized)
		w.Write(json.RawMessage(`"error": "Expired token."`))
		return
	}

	newToken, err := auth.MakeRefreshToken()
	if err != nil {
		log.Printf("POST /api/refresh: Error creating new refresh token: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(json.RawMessage(`"error": "Error creating new refresh token"`))
		return
	}

	refreshTokenParams := database.CreateRefreshTokenParams{
		Token:     newToken,
		UserID:    token.UserID,
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	}

	err = cfg.queries.CreateRefreshToken(r.Context(), refreshTokenParams)

	if err != nil {
		log.Printf("POST /api/refresh: Error saving new refresh token in database: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(json.RawMessage(`"error": "Error creating new refresh token"`))
		return
	}

	res := &struct {
		Token string `json:"token"`
	}{
		newToken,
	}
	rawRes, err := json.Marshal(res)
	if err != nil {
		log.Printf("POST /api/refresh: Error marshalling json response: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(json.RawMessage(`"error": "Error marshalling json response"`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(rawRes)
	return
}
