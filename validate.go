package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func validateChirpHandler(w http.ResponseWriter, req *http.Request) {

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

	type returnVals struct {
		// the key will be the name of struct field unless you give it an explicit JSON tag
		CleanedBody string `json:"cleaned_body"`
	}

	respondWithJSON(w, http.StatusOK, returnVals{
		CleanedBody: bodyClean,
	})

}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	if code > 499 {
		log.Printf("Responding with 5XX error: %s", msg)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, errorResponse{
		Error: msg,
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
}
