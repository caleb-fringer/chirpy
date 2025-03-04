package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/caleb-fringer/chirpy/internal/auth"
	"github.com/google/uuid"
)

const UPGRADE_EVENT = "user.upgraded"

func (cfg *apiConfig) subscribe(w http.ResponseWriter, r *http.Request) {
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write(json.RawMessage(`{"error": "Please provide an API key in the Authorization header"}`))
		return
	}

	if apiKey != cfg.polkaKey {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write(json.RawMessage(`{"error": "Invalid API key"}`))
		return
	}

	reqBody := &struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}{}

	rawReqBody, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("POST /api/polka/webhooks: Error reading request body: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(json.RawMessage(`{"error": "Bad request body.}`))
		return
	}

	err = json.Unmarshal(rawReqBody, reqBody)
	if err != nil {
		log.Printf("POST /api/polka/webhooks: Error unmarshalling request body: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(json.RawMessage(`{"error": "Bad request body.}`))
		return
	}

	// Only respond to upgrade events
	if reqBody.Event != UPGRADE_EVENT {
		w.WriteHeader(http.StatusNoContent)
		w.Write(nil)
		return
	}

	err = cfg.queries.UpgradeToChirpyRed(r.Context(), uuid.MustParse(reqBody.Data.UserID))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write(nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	w.Write(nil)
	return
}
