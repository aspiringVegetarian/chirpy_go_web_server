package database

import (
	"fmt"
	"log"
)

type User struct {
	// the key will be the name of struct field unless you give it an explicit JSON tag
	ID             int    `json:"id"`
	Email          string `json:"email"`
	HashedPassword string `json:"hashed_password"`
}

func (db *DB) CreateUser(email string, password string) (User, error) {
	if _, exists := db.UserIDLookup(email); exists {
		log.Printf("Email is already registered")
		return User{}, fmt.Errorf("email already registered")
	}
	id := len(db.Data.Users) + 1
	user := User{
		ID:             id,
		Email:          email,
		HashedPassword: password,
	}
	db.Data.Users[id] = user
	err := db.writeDB(db.Data)
	if err != nil {
		log.Printf("Failed to write new user to database")
		return User{}, err
	}
	return user, nil

}

func (db *DB) UpdateUser(id int, email, password string) (User, error) {

	if _, exist := db.Data.Users[id]; !exist {
		log.Printf("Attempted to update ser ID %v, which does not exist", id)
		return User{}, fmt.Errorf("Attempted to update ser ID %v, which does not exist", id)
	}

	updatedUser := User{
		ID:             id,
		Email:          email,
		HashedPassword: password,
	}

	db.Data.Users[id] = updatedUser
	err := db.writeDB(db.Data)
	if err != nil {
		log.Printf("Failed to write updated user to database")
		return User{}, err
	}
	return updatedUser, nil

}

func (db *DB) UserIDLookup(email string) (int, bool) {
	for id, val := range db.Data.Users {
		if val.Email == email {
			return id, true
		}
	}
	return 0, false
}
