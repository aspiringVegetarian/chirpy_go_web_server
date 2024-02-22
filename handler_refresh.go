package main

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type RefreshToken struct {
	// the key will be the name of struct field unless you give it an explicit JSON tag
	Token string `json:"token"`
}

func (cfg *apiConfig) postRefreshHandler(w http.ResponseWriter, req *http.Request) {

	header := req.Header.Get("Authorization")

	tokenString := strings.TrimPrefix(header, "Bearer ")

	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.jwtSecret), nil
	})

	if err != nil {
		log.Printf("Error parsing token: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Bad Token")
		return
	}

	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		log.Printf("Could not extract issuer from token claims: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Could not extract issuer from token")
		return
	}

	if issuer != "chirpy-refresh" {
		log.Printf("Invalid token issuer: %s", issuer)
		respondWithError(w, http.StatusUnauthorized, "Invalid token issuer")
		return
	}

	if _, exists := cfg.revokedTokens[tokenString]; exists {
		log.Printf("Refresh token has been revoked and is no longer valid: %s", tokenString)
		respondWithError(w, http.StatusUnauthorized, "Invalid refresh token. This token has been revoked.")
		return
	}

	idString, err := token.Claims.GetSubject()
	if err != nil {
		log.Printf("Could not extract subject from token claims: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Could not extract subject from token")
		return
	}

	timeNow := time.Now().UTC()

	access_Claims := &jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		IssuedAt:  jwt.NewNumericDate(timeNow),
		ExpiresAt: jwt.NewNumericDate(timeNow.Add(time.Hour)),
		Subject:   idString,
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, access_Claims)
	signedAccessToken, err := accessToken.SignedString([]byte(cfg.jwtSecret))
	if err != nil {
		log.Printf("Error generating a signed string of the JWT with error: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	respondWithJSON(w, http.StatusOK, RefreshToken{Token: signedAccessToken})

}

func (cfg *apiConfig) postRevokeTokenHandler(w http.ResponseWriter, req *http.Request) {

	header := req.Header.Get("Authorization")

	tokenString := strings.TrimPrefix(header, "Bearer ")

	cfg.revokedTokens[tokenString] = time.Now().UTC()

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	w.WriteHeader(http.StatusOK)

}
