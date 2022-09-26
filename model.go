package main

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	numGuesses = 6 // Maximum number of guesses you can make
	numChars   = 5 // Word size in characters
)

var _ tea.Model = (*model)(nil)

// Represents the state of a key
type keyState int

// Sent when the status line should be reset
type msgResetStatus struct{}

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

func (m *model) Init() tea.Cmd {
	m.keyStates = make(map[byte]keyState, 26)
	return m.withDb(func(db *db) {
		m.score = db.score()
		m.reset()
	})
}

// Sets the status message, and returns a tea.Cmd that restores the
// default status message after a delay
func (m *model) setStatus(msg string, duration time.Duration) tea.Cmd {
	m.status = msg
	if duration > 0 {
		m.statusPending++
		return tea.Tick(duration, func(time.Time) tea.Msg {
			return msgResetStatus{}
		})
	}
	return nil
}

// withDb runs a function in the context of the database. The database is
// automatically saved at the end.
func (m *model) withDb(f func(db *db)) tea.Cmd {
	db, err := loadDb()
	if err != nil {
		return m.reportError(err, "Error loading database.")
	}
	f(db)
	if err := db.save(); err != nil {
		return m.reportError(err, "Error saving database.")
	}
	return nil
}

// Stores the given error and prints a message to the status line
func (m *model) reportError(err error, msg string) tea.Cmd {
	m.errors = append(m.errors, err)
	return m.setStatus(msg, 3*time.Second)
}

// Immediately resets the status message to its default value
func (m *model) resetStatus() {
	m.status = fmt.Sprintf("Score: %d", m.score)
}

func (m *model) reset() {
	// Unlock and reset the grid
	m.gameOver = false
	m.gridCol = 0
	m.gridRow = 0
	// Clear the key state
	for k := range m.keyStates {
		delete(m.keyStates, k)
	}
	// Set the puzzle word
	word := getWord()
	copy(m.word[:], word)
	// Reset the status message
	m.resetStatus()
}
