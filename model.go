package main

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	numGuesses = 6 // Maximum number of guesses you can make
	numChars   = 5 // Word size in characters

	keyStateUnselected keyState = iota
	keyStateAbsent
	keyStatePresent
	keyStateCorrect

	colorPrimary   = lipgloss.Color("#d7dadc")
	colorSecondary = lipgloss.Color("#626262")
	colorSeparator = lipgloss.Color("#9c9c9c")
	colorYellow    = lipgloss.Color("#b59f3b")
	colorGreen     = lipgloss.Color("#538d4e")
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

// Exits the program
func (*model) doExit() tea.Cmd {
	return tea.Quit
}

// Deletes the last character in the current word
func (m *model) doDeleteChar() tea.Cmd {
	if !m.gameOver && m.gridCol > 0 {
		m.gridCol--
	}
	return nil
}

// Called when the user has guessed the word correctly
func (m *model) doWin() tea.Cmd {
	m.gameOver = true
	return tea.Sequentially(
		m.withDb(func(db *db) {
			db.addWin(m.gridRow)
			m.score = db.score()
		}),
		m.setStatus("You win!", 0),
	)
}

// Returns the appropriate dark mode color for the given key state
func (s keyState) color() lipgloss.Color {
	switch s {
	case keyStateUnselected:
		return colorPrimary
	case keyStateAbsent:
		return colorSecondary
	case keyStatePresent:
		return colorYellow
	case keyStateCorrect:
		return colorGreen
	default:
		panic("invalid key status")
	}
}
