package database

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/go-chi/chi/v5"
)

type Chirp struct {
	// the key will be the name of struct field unless you give it an explicit JSON tag
	ID   int    `json:"id"`
	Body string `json:"body"`
}

type User struct {
	// the key will be the name of struct field unless you give it an explicit JSON tag
	ID    int    `json:"id"`
	Email string `json:"email"`
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
	Users  map[int]User  `json:"users"`
}

type DB struct {
	path string
	mux  *sync.RWMutex
	data DBStructure
}

// NewDB creates a new database connection
// and creates the database file if it doesn't exist
func NewDB(path string) (*DB, error) {
	newDB := DB{
		path: path,
		mux:  &sync.RWMutex{},
	}

	err := newDB.ensureDB()
	if err != nil {
		log.Printf("Failed to create new database")
		return &newDB, err
	}

	newDB.data, err = newDB.loadDB()
	if err != nil {
		log.Printf("Failed to load new database")
		return &newDB, err
	}

	return &newDB, nil
}

// ensureDB creates a new database file if it doesn't exist
func (db *DB) ensureDB() error {
	// check path for existing db file
	_, err := os.Stat(db.path)
	if errors.Is(err, os.ErrNotExist) {
		// if not, create new data of type DBStructure and then write data to path
		log.Printf("Database does not exist at path: %v", db.path)
		log.Printf("Creating new database at path: %v", db.path)

		db.data = DBStructure{
			Chirps: make(map[int]Chirp),
			Users:  make(map[int]User),
		}

		err := db.writeDB(db.data)
		if err != nil {
			log.Printf("Failed to write new database")
			return err
		}
		return nil
	}
	return err
}

// loadDB reads the database file into memory
func (db *DB) loadDB() (DBStructure, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	data, err := os.ReadFile(db.path)
	if err != nil {
		log.Printf("Failed to read database")
		return DBStructure{}, err
	}

	err = json.Unmarshal(data, &db.data)
	if err != nil {
		log.Printf("Failed to unmarshal data")
		return DBStructure{}, err
	}

	return db.data, nil
}

// writeDB writes the database file to disk
func (db *DB) writeDB(dbStructure DBStructure) error {
	data, err := json.Marshal(dbStructure)
	if err != nil {
		log.Printf("Failed to marshal data")
		return err
	}

	db.mux.Lock()
	defer db.mux.Unlock()

	err = os.WriteFile(db.path, data, 0600)
	if err != nil {
		log.Printf("Failed to write new database")
		return err
	}

	return nil
}

// CreateChirp creates a new chirp and saves it to disk
func (db *DB) CreateChirp(body string) (Chirp, error) {
	id := len(db.data.Chirps) + 1
	chirp := Chirp{
		ID:   id,
		Body: body,
	}
	db.data.Chirps[id] = chirp
	err := db.writeDB(db.data)
	if err != nil {
		log.Printf("Failed to write new database")
		return Chirp{}, err
	}
	return chirp, nil

}

func (db *DB) CreateUser(email string) (User, error) {
	id := len(db.data.Users) + 1
	user := User{
		ID:    id,
		Email: email,
	}
	db.data.Users[id] = user
	err := db.writeDB(db.data)
	if err != nil {
		log.Printf("Failed to write new database")
		return User{}, err
	}
	return user, nil

}

// GetChirps returns all chirps in the database
func (db *DB) GetChirps() ([]Chirp, error) {
	chirps := make([]Chirp, 0, len(db.data.Chirps))

	for _, value := range db.data.Chirps {
		chirps = append(chirps, value)
	}

	sort.Slice(chirps, func(i, j int) bool { return chirps[i].ID < chirps[j].ID })

	return chirps, nil
}

func (db *DB) GetChirpsHandler(w http.ResponseWriter, req *http.Request) {

	chirps, err := db.GetChirps()
	if err != nil {
		log.Printf("Failed to get chirps with error: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Couldn't get chirps")
		return
	}

	respondWithJSON(w, http.StatusOK, chirps)

}

func (db *DB) GetChirpHandler(w http.ResponseWriter, req *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(req, "chirpID"))
	if err != nil {
		log.Printf("Failed to get chirp ID from request with error: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Couldn't get chirp ID")
		return
	}

	chirp, exists := db.data.Chirps[id]
	if !exists {
		log.Printf("Chirp ID %v does not exist", id)
		respondWithError(w, http.StatusNotFound, "Chirp ID requested does not exist")
		return
	}

	respondWithJSON(w, http.StatusOK, chirp)

}

func (db *DB) PostUserHandler(w http.ResponseWriter, req *http.Request) {

	type parameters struct {
		// these tags indicate how the keys in the JSON should be mapped to the struct fields
		// the struct fields must be exported (start with a capital letter) if you want them parsed
		Email string `json:"email"`
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

	newUser, err := db.CreateUser(params.Email)
	if err != nil {
		log.Printf("Failed to create new chirp with error: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Couldn't create new chirp")
		return
	}

	respondWithJSON(w, http.StatusCreated, newUser)

}

func (db *DB) PostChirpHandler(w http.ResponseWriter, req *http.Request) {

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

	newChirp, err := db.CreateChirp(bodyClean)
	if err != nil {
		log.Printf("Failed to create new chirp with error: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Couldn't create new chirp")
		return
	}

	respondWithJSON(w, http.StatusCreated, newChirp)

}

func (db *DB) DatabaseResetHandler(w http.ResponseWriter, req *http.Request) {
	db.data = DBStructure{
		Chirps: make(map[int]Chirp),
		Users:  make(map[int]User),
	}

	err := db.writeDB(db.data)
	if err != nil {
		log.Printf("Failed to write new database")

	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	w.WriteHeader(http.StatusOK)

	w.Write([]byte("Database has been reset"))
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
