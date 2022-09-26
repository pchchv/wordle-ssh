package main

import (
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	tea "github.com/charmbracelet/bubbletea"
)

const envHostKey = "_CLIDLE_HOSTKEY"

var (
	pathClidle  = filepath.Join(xdg.DataHome, "clidle")
	pathHostKey = filepath.Join(pathClidle, "hostkey")
	teaOptions  = []tea.ProgramOption{tea.WithAltScreen(), tea.WithOutput(os.Stderr)}
	pathDb      = filepath.Join(pathClidle, "db.json")
)

func main() {}

func server(addr string) {

}
