package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/adrg/xdg"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/wish"
	bm "github.com/charmbracelet/wish/bubbletea"
	lm "github.com/charmbracelet/wish/logging"
	"github.com/gliderlabs/ssh"
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
	withHostKey := wish.WithHostKeyPath(pathHostKey)
	if pem, ok := os.LookupEnv(envHostKey); ok {
		withHostKey = wish.WithHostKeyPEM([]byte(pem))
	}
	server, err := wish.NewServer(
		wish.WithAddress(addr),
		wish.WithIdleTimeout(30*time.Minute),
		wish.WithMiddleware(
			bm.Middleware(func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
				pty, _, active := s.Pty()
				if !active {
					log.Printf("no active terminal, skipping")
					return nil, nil
				}
				model := &model{
					width:  pty.Window.Width,
					height: pty.Window.Height,
				}
				return model, teaOptions
			}),
			lm.Middleware(),
		),
		withHostKey,
	)
	if err != nil {
		log.Fatalf("could not create server: %s", err)
	}
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("starting server: %s", server.Addr)
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("server returned an error: %s", err)
		}
	}()
	<-done
	log.Println("stopping server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("could not shutdown server gracefully: %s", err)
	}
}
