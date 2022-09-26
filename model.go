package main

import tea "github.com/charmbracelet/bubbletea"

const (
	numGuesses = 6 // Maximum number of guesses you can make
	numChars   = 5 // Word size in characters
)

var _ tea.Model = (*model)(nil)

type model struct {
	score         int
	word          [numChars]byte
	gameOver      bool
	errors        []error
	keyStates     map[byte]keyState
	status        string
	statusPending int
	height        int
	width         int
	grid          [numGuesses][numChars]byte
	gridRow       int
	gridCol       int
}

// Represents the state of a key
type keyState int
