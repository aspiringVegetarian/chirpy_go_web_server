package database

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"sync"
)

type DBStructure struct {
	Chirps map[int]Chirp   `json:"chirps"`
	Users  map[int]User `json:"users"`
}

type DB struct {
	path string
	mux  *sync.RWMutex
	Data DBStructure
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

	newDB.Data, err = newDB.loadDB()
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

		db.Data = DBStructure{
			Chirps: make(map[int]Chirp),
			Users:  make(map[int]User),
		}

		err := db.writeDB(db.Data)
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

	err = json.Unmarshal(data, &db.Data)
	if err != nil {
		log.Printf("Failed to unmarshal data")
		return DBStructure{}, err
	}

	return db.Data, nil
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

func (db *DB) DatabaseResetHandler(w http.ResponseWriter, req *http.Request) {
	err := os.Remove(db.path)
	if err != nil {
		log.Fatal(err)
	}

	err = db.ensureDB()
	if err != nil {
		log.Fatal(err)
	}

	db.Data, err = db.loadDB()
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	w.WriteHeader(http.StatusOK)

	w.Write([]byte("Database has been reset"))
}
