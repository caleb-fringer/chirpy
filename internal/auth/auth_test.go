package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

const pw = "password"

var hash string
var id = uuid.New()
var secret = "letmein!"
var token string

func TestHashPassword(t *testing.T) {
	var err error
	hash, err = HashPassword(pw)
	if err != nil {
		t.Fatalf("Error hashing password: %v\n", err)
	}
}

func TestCheckPasswordHash(t *testing.T) {
	err := CheckPasswordHash(pw, hash)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
}

func TestMakeJWT(t *testing.T) {
	var err error
	token, err = MakeJWT(id, secret, 5*time.Minute)
	if err != nil {
		t.Fatalf("Error making JWT: %v\n", err)
	}
}

func TestValidateJWT(t *testing.T) {
	retrievedId, err := ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("Error validating JWT: %v\n", err)
	}
	if id != retrievedId {
		t.Fatalf("IDs don't match: expected: %s got: %s\n", id, retrievedId)
	}
}

func TestGetBearerToken(t *testing.T) {
	tokenStr := "asfeqjwrlkelkj1234132"
	header := http.Header{}
	header.Add("Authorization", "Bearer "+tokenStr)

	token, err := GetBearerToken(header)
	if err != nil {
		t.Fatalf("Error getting bearer token: %v\n", err)
	}

	if token != tokenStr {
		t.Fatalf("Extracted token does not match expected value: expected %s got %s", tokenStr, token)
	}
}
