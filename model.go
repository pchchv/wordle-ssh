package main

import (
	"bytes"
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

// Accepts the current word
func (m *model) doAcceptWord() tea.Cmd {
	if m.gameOver {
		return nil
	}
	// Only accept a word if it is complete
	if m.gridCol != numChars {
		return m.setStatus("Your guess must be a 5-letter word.", 1*time.Second)
	}
	// Check if the input word is valid
	word := m.grid[m.gridRow]
	if !isWord(string(word[:])) {
		return m.setStatus("That's not a valid word.", 1*time.Second)
	}
	// Update the state of the used letters
	success := true
	for i := 0; i < numChars; i++ {
		key := word[i]
		keyStatus := keyStateAbsent
		if key == m.word[i] {
			keyStatus = keyStateCorrect
		} else {
			success = false
			if bytes.IndexByte(m.word[:], key) != -1 {
				keyStatus = keyStatePresent
			}
		}
		if m.keyStates[key] < keyStatus {
			m.keyStates[key] = keyStatus
		}
	}
	// Move to the next row
	m.gridRow++
	m.gridCol = 0
	// Check if the game is over
	if success {
		return m.doWin()
	} else if m.gridRow == numGuesses {
		return m.doLoss()
	}
	return nil
}

// Adds one input character to the current word
func (m *model) doAcceptChar(ch rune) tea.Cmd {
	// Only accept a character if the current word is incomplete
	if m.gameOver || !(m.gridRow < numGuesses && m.gridCol < numChars) {
		return nil
	}
	ch = toAsciiUpper(ch)
	if isAsciiUpper(ch) {
		m.grid[m.gridRow][m.gridCol] = byte(ch)
		m.gridCol++
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

// Called when the user has used up all their guesses
func (m *model) doLoss() tea.Cmd {
	m.gameOver = true
	msg := fmt.Sprintf("The word was %s. Better luck next time!", string(m.word[:]))
	return tea.Sequentially(
		m.withDb(func(db *db) {
			db.addLoss()
			m.score = db.score()
		}),
		m.setStatus(msg, 0),
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

// Converts a rune to uppercase if it is between A-Z
func toAsciiUpper(r rune) rune {
	if 'a' <= r && r <= 'z' {
		r -= 'a' - 'A'
	}
	return r
}

// Checks if a rune is between A-Z
func isAsciiUpper(r rune) bool {
	return 'A' <= r && r <= 'Z'
}
