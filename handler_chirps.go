package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

func (cfg *apiConfig) getChirpsHandler(w http.ResponseWriter, req *http.Request) {

	chirps, err := cfg.chirpyDatabase.GetChirps()
	if err != nil {
		log.Printf("Failed to get chirps with error: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Couldn't get chirps")
		return
	}

	respondWithJSON(w, http.StatusOK, chirps)

}

func (cfg *apiConfig) getChirpHandler(w http.ResponseWriter, req *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(req, "chirpID"))
	if err != nil {
		log.Printf("Failed to get chirp ID from request with error: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Couldn't get chirp ID")
		return
	}

	chirp, exists := cfg.chirpyDatabase.Data.Chirps[id]
	if !exists {
		log.Printf("Chirp ID %v does not exist", id)
		respondWithError(w, http.StatusNotFound, "Chirp ID requested does not exist")
		return
	}

	respondWithJSON(w, http.StatusOK, chirp)

}

func (cfg *apiConfig) postChirpHandler(w http.ResponseWriter, req *http.Request) {

	type parameters struct {
		// these tags indicate how the keys in the JSON should be mapped to the struct fields
		// the struct fields must be exported (start with a capital letter) if you want them parsed
		Body string `json:"body"`
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

	if len(params.Body) > 140 {
		log.Printf("Chirp is longer than 140 characters. Chirp length: %v", len(params.Body))
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	bodySplit := strings.Split(params.Body, " ")
	badWords := map[string]bool{
		"kerfuffle": true,
		"sharbert":  true,
		"fornax":    true}
	for i, word := range bodySplit {
		if badWords[strings.ToLower(word)] {
			bodySplit[i] = "****"
		}
	}

	bodyClean := strings.Join(bodySplit, " ")

	newChirp, err := cfg.chirpyDatabase.CreateChirp(bodyClean)
	if err != nil {
		log.Printf("Failed to create new chirp with error: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Couldn't create new chirp")
		return
	}

	respondWithJSON(w, http.StatusCreated, newChirp)

}
