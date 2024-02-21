package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	// the key will be the name of struct field unless you give it an explicit JSON tag
	ID    int    `json:"id"`
	Email string `json:"email"`
}

func (cfg *apiConfig) postUserHandler(w http.ResponseWriter, req *http.Request) {

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

	if params.Password == "" {
		log.Printf("User did not enter a password")
		respondWithError(w, http.StatusBadRequest, "Enter a password")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)

	if err != nil {
		log.Printf("Failed to create new user with error: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Could not hash password")
		return
	}

	newUser, err := cfg.chirpyDatabase.CreateUser(params.Email, string(hashedPassword))

	if err != nil {
		log.Printf("Failed to create new user with error: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Email already registered")
		return
	}

	respondWithJSON(w, http.StatusCreated, User{ID: newUser.ID, Email: newUser.Email})

}

func (cfg *apiConfig) putUserHandler(w http.ResponseWriter, req *http.Request) {

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

	log.Print(token)

	idString, err := token.Claims.GetSubject()
	if err != nil {
		log.Printf("Could not extract subject from token claims: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Could not extract subject from token")
		return
	}

	id, err := strconv.Atoi(idString)
	if err != nil {
		log.Printf("Could not convert id string to int: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Could not convert id string to int")
		return
	}

	if params.Password == "" {
		log.Printf("User did not enter a password")
		respondWithError(w, http.StatusBadRequest, "Enter a password")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)

	if err != nil {
		log.Printf("Failed to generate hashed password with error: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Could not hash password")
		return
	}

	newUser, err := cfg.chirpyDatabase.UpdateUser(id, params.Email, string(hashedPassword))

	if err != nil {
		log.Printf("Failed to update user with error: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Could not update user")
		return
	}

	respondWithJSON(w, http.StatusOK, User{ID: newUser.ID, Email: newUser.Email})

}
