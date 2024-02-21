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
	ID    int    `json:"id"`
	Email string `json:"email"`
	Token string `json:"token"`
}

func (cfg *apiConfig) postLoginHandler(w http.ResponseWriter, req *http.Request) {

	type parameters struct {
		// these tags indicate how the keys in the JSON should be mapped to the struct fields
		// the struct fields must be exported (start with a capital letter) if you want them parsed
		Email      string `json:"email"`
		Password   string `json:"password"`
		Expiration int    `json:"expires_in_seconds"`
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

	if params.Expiration == 0 || params.Expiration > 24*60*60 {
		params.Expiration = 24 * 60 * 60
	}

	claims := &jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * time.Duration(params.Expiration))),
		Subject:   strconv.Itoa(user.ID),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(cfg.jwtSecret))
	if err != nil {
		log.Printf("Error generating a signed string of the JWT with error: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	respondWithJSON(w, http.StatusOK, UserToken{ID: user.ID, Email: user.Email, Token: ss})

}
