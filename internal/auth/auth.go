package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return "", fmt.Errorf("Error hashing password: %v", err)
	}
	return string(hashed), nil
}

func CheckPasswordHash(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return err
	}
	return nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string) (string, error) {
	issueTime := time.Now()
	expireTime := issueTime.Add(time.Hour)
	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(issueTime),
		ExpiresAt: jwt.NewNumericDate(expireTime),
		Subject:   userID.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(tokenSecret))

	if err != nil {
		log.Printf("Error signing JWT: %v\n", err)
		return "", err
	}

	return signedToken, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	})

	if err != nil {
		log.Printf("Error parsing JWT: %v\n", err)
		return uuid.UUID{}, err
	}

	subject, err := token.Claims.GetSubject()
	if err != nil {
		log.Printf("Error getting token subject: %v\n", err)
		return uuid.UUID{}, err
	}

	id := uuid.MustParse(subject)
	return id, nil
}

type HeaderNotFoundError struct {
	Header string
}

func (e HeaderNotFoundError) Error() string {
	return fmt.Sprintf("%s header not found.", e.Header)
}

type WrongAuthorizationSchemeError struct {
}

func (e WrongAuthorizationSchemeError) Error() string {
	return "Wrong authorization scheme. Please use the Bearer scheme"
}

func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", HeaderNotFoundError{Header: "Authorization"}
	}

	token, ok := strings.CutPrefix(authHeader, "Bearer ")
	if !ok {
		return "", fmt.Errorf("Authorization header is not using Bearer scheme")
	}

	return token, nil
}

const REFRESH_TOKEN_LENGTH = 32

func MakeRefreshToken() (string, error) {
	token := make([]byte, REFRESH_TOKEN_LENGTH)
	_, err := rand.Read(token)
	if err != nil {
		return "", fmt.Errorf("Error creating refresh token: %v", err)
	}

	return hex.EncodeToString(token), nil
}
