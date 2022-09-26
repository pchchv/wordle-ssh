package main

// db is the file where the game statistics are stored.
type db struct {
	// Guesses stores the win statistics of each game. Guesses[0] is the number
	// of games that were lost, Guesses[1] is the number of games that were won
	// with 1 guess, etc.
	Guesses [numGuesses + 1]int `json:"guesses"`
}
