package main

import (
	"encoding/json"
	"os"

	"github.com/pkg/errors"
)

// File where the game statistics are stored
type db struct {
	// Guesses stores the win statistics of each game. Guesses[0] is the number
	// of games that were lost, Guesses[1] is the number of games that were won
	// with 1 guess, etc.
	Guesses [numGuesses + 1]int `json:"guesses"`
}

// Reads the database from dbPath
func loadDb() (*db, error) {
	file, err := os.Open(pathDb)
	if err != nil {
		if os.IsNotExist(err) {
			return &db{}, nil
		}
		return nil, errors.Wrap(err, "could not find database")
	}
	var db db
	if err := json.NewDecoder(file).Decode(&db); err != nil {
		return nil, errors.Wrap(err, "could not read from database")
	}
	return &db, nil
}
