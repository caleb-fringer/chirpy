package main

import (
	"log"
	"strings"
)

var profaneWords = map[string]struct{}{
	"kerfuffle": {},
	"sharbert":  {},
	"fornax":    {},
	"farking":   {},
}

func validateChirp(chirp string) (string, bool) {
	if len(chirp) > 140 {
		return "", false
	}

	censored := censorProfanity(chirp)
	return censored, true
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
