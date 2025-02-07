package auth

import "testing"

const pw = "password"

var hash string

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
