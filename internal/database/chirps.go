package database

import (
	"log"
	"sort"
)

type Chirp struct {
	// the key will be the name of struct field unless you give it an explicit JSON tag
	ID   int    `json:"id"`
	Body string `json:"body"`
}

// CreateChirp creates a new chirp and saves it to disk
func (db *DB) CreateChirp(body string) (Chirp, error) {
	id := len(db.Data.Chirps) + 1
	chirp := Chirp{
		ID:   id,
		Body: body,
	}
	db.Data.Chirps[id] = chirp
	err := db.writeDB(db.Data)
	if err != nil {
		log.Printf("Failed to write new database")
		return Chirp{}, err
	}
	return chirp, nil

}

// GetChirps returns all chirps in the database
func (db *DB) GetChirps() ([]Chirp, error) {
	chirps := make([]Chirp, 0, len(db.Data.Chirps))

	for _, value := range db.Data.Chirps {
		chirps = append(chirps, value)
	}

	sort.Slice(chirps, func(i, j int) bool { return chirps[i].ID < chirps[j].ID })

	return chirps, nil
}
