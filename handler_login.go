package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserToken struct {
	// the key will be the name of struct field unless you give it an explicit JSON tag
	ID           int    `json:"id"`
	Email        string `json:"email"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

func (cfg *apiConfig) postLoginHandler(w http.ResponseWriter, req *http.Request) {

	type parameters struct {
		// these tags indicate how the keys in the JSON should be mapped to the struct fields
		// the struct fields must be exported (start with a capital letter) if you want them parsed
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		// an error will be thrown if the JSON is invalid or has the wrong types
		// any missing fields will simply have their values in the struct set to their zero value
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	id, exists := cfg.chirpyDatabase.UserIDLookup(params.Email)

	if !exists {
		log.Printf("Email entered, %s, does not match a registered email", params.Email)
		respondWithError(w, http.StatusBadRequest, "Email entered does not match a registered email")
		return
	}

	user := cfg.chirpyDatabase.Data.Users[id]

	passwordCheckErr := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(params.Password))

	if passwordCheckErr != nil {
		log.Printf("Password entered does not match stored password for this email address: %s", params.Email)
		respondWithError(w, http.StatusUnauthorized, "Password entered does not match stored password for this email address")
		return
	}

	timeNow := time.Now().UTC()

	access_Claims := &jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		IssuedAt:  jwt.NewNumericDate(timeNow),
		ExpiresAt: jwt.NewNumericDate(timeNow.Add(time.Hour)),
		Subject:   strconv.Itoa(user.ID),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, access_Claims)
	signedAccessToken, err := accessToken.SignedString([]byte(cfg.jwtSecret))
	if err != nil {
		log.Printf("Error generating a signed string of the JWT with error: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	refreshExpiration := time.Duration(time.Hour * 24 * 60)

	refreshClaims := &jwt.RegisteredClaims{
		Issuer:    "chirpy-refresh",
		IssuedAt:  jwt.NewNumericDate(timeNow),
		ExpiresAt: jwt.NewNumericDate(timeNow.Add(refreshExpiration)),
		Subject:   strconv.Itoa(user.ID),
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	signedRefreshToken, err := refreshToken.SignedString([]byte(cfg.jwtSecret))
	if err != nil {
		log.Printf("Error generating a signed string of the JWT with error: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	respondWithJSON(w, http.StatusOK, UserToken{ID: user.ID, Email: user.Email, Token: signedAccessToken, RefreshToken: signedRefreshToken})

}
