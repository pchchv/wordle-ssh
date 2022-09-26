package main

import (
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

// Stores the given error and prints a message to the status line
func (m *model) reportError(err error, msg string) tea.Cmd {
	m.errors = append(m.errors, err)
	return m.setStatus(msg, 3*time.Second)
}
