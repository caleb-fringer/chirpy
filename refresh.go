package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/caleb-fringer/chirpy/internal/auth"
	"github.com/caleb-fringer/chirpy/internal/database"
	"github.com/google/uuid"
)

type refreshResponse struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
}

func (cfg *apiConfig) refresh(w http.ResponseWriter, r *http.Request) {
	tokenStr, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("POST /api/refresh: Error extracting refresh token from request headers: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
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
		w.WriteHeader(http.StatusUnauthorized)
		w.Write(json.RawMessage(`"error": "Expired token."`))
		return
	}

	if token.RevokedAt.Valid {
		log.Printf("Warning: Someone attempted to use a revoked refresh token: token %s\n", tokenStr)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write(json.RawMessage(`"error": "Revoked token."`))
		return
	}

	refreshTokenStr, err := auth.MakeRefreshToken()
	if err != nil {
		log.Printf("POST /api/refresh: Error creating new refresh token: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(json.RawMessage(`"error": "Error creating new refresh token"`))
		return
	}

	refreshTokenParams := database.CreateRefreshTokenParams{
		Token:     refreshTokenStr,
		UserID:    token.UserID,
		ExpiresAt: time.Now().UTC().Add(60 * 24 * time.Hour),
	}

	refreshToken, err := cfg.queries.CreateRefreshToken(r.Context(), refreshTokenParams)

	if err != nil {
		log.Printf("POST /api/refresh: Error saving new refresh token in database: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(json.RawMessage(`"error": "Error creating new refresh token"`))
		return
	}

	newAuthToken, err := auth.MakeJWT(token.UserID, cfg.secretKey)
	if err != nil {
		log.Printf("POST /api/refresh: Error creating new JWT: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(json.RawMessage(`"error": "Error creating new authorization token"`))
		return
	}

	user, err := cfg.queries.GetUserByID(r.Context(), refreshToken.UserID)
	if err != nil {
		log.Printf("POST /api/refresh: Error getting user email from UUID: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(json.RawMessage(`"error": "Error creating new authorization token"`))
		return
	}

	res := &refreshResponse{
		user.ID,
		refreshToken.CreatedAt,
		refreshToken.UpdatedAt,
		user.Email,
		newAuthToken,
		refreshToken.Token,
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

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	tokenStr, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("POST /api/revoke: Error extracting authorization token: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(json.RawMessage(`"error": "Please provide an authorization token in the authorization header."`))
		return
	}

	err = cfg.queries.RevokeRefreshToken(r.Context(), tokenStr)
	if err != nil {
		log.Printf("POST /api/revoke: Error revoking refresh token in DB: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(json.RawMessage(`"error": "Database error"`))
		return
	}

	w.WriteHeader(http.StatusNoContent)
	w.Write(nil)
	return
}
