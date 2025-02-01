package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type response struct {
	Body string `json:"cleaned_body"`
}

var profaneWords = map[string]struct{}{
	"kerfuffle": {},
	"sharbert":  {},
	"fornax":    {},
}

func validateChirp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	type reqBody struct {
		Body string
	}

	decoder := json.NewDecoder(r.Body)
	body := reqBody{}

	err := decoder.Decode(&body)
	if err != nil {
		log.Printf("POST /api/validate_chirp: Error decoding HTTP request body: %s\n", err)
		w.WriteHeader(500)
		w.Write(json.RawMessage(`{"error": "Error decoding HTTP request."}`))
		return
	}

	if len(body.Body) > 140 {
		w.WriteHeader(400)
		w.Write(json.RawMessage(`{"error":"Chirp is too long"}`))
		return
	}

	response := response{censorProfanity(body.Body)}
	rawResBody, err := json.Marshal(response)

	if err != nil {
		log.Printf("POST /api/validate_chirp: Error marshalling censored Chirp: %s\n", err)
		w.WriteHeader(500)
		w.Write(json.RawMessage(`{"error": "Server error."}`))
		return
	}

	w.WriteHeader(200)
	w.Write(rawResBody)
	return
}

func censorProfanity(in string) string {
	words := strings.Split(in, " ")
	for i, word := range words {
		log.Printf("Checking %s for profanity...\n", word)
		_, ok := profaneWords[strings.ToLower(word)]
		if ok {
			words[i] = strings.Repeat("*", 4)
		} else {
			words[i] = word
		}
	}

	return strings.Join(words, " ")
}
