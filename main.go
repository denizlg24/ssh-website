package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/logging"
	"github.com/muesli/termenv"
)

const portrait = `........................................
........................................
.......###################..............
.......##---------------%###............
.......##------------------##...........
.......####----#######------%##.........
.........##----#########----%##%%.......
.........##----#########----%##%%.......
.........##----#########----%##%%.......
.........##----#########----%##%%.......
.........##----#######------%##%%.......
.......##------------------####%%.......
.......##---------------#####%%%%.......
.......######################%%%%.......
.......####################%%%%.........
.........%%%%%%%%%%%%%%%%%%%%-..........
........................................
........................................
`

func main() {
	if os.Getenv("PREVIEW") == "1" {
		r := lipgloss.NewRenderer(os.Stdout, termenv.WithProfile(termenv.TrueColor))
		fmt.Print(render(r, 80))
		return
	}

	host := envOr("HOST", "0.0.0.0")
	port := envOr("PORT", "22")
	keyPath := envOr("HOST_KEY_PATH", ".ssh/host_key")

	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(keyPath),
		wish.WithMiddleware(
			displayMiddleware(),
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Fatalf("failed to start server: %v", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Printf("ssh server listening on %s:%s", host, port)

	go func() {
		if err := s.ListenAndServe(); err != nil {
			log.Printf("server stopped: %v", err)
		}
	}()

	<-done
	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func displayMiddleware() wish.Middleware {
	return func(next ssh.Handler) ssh.Handler {
		return func(sess ssh.Session) {
			width := 80
			pty, _, ok := sess.Pty()
			if ok && pty.Window.Width > 0 {
				width = pty.Window.Width
			}

			renderer := lipgloss.NewRenderer(sess, termenv.WithProfile(termenv.TrueColor))
			output := render(renderer, width)
			fmt.Fprint(sess, output)
			next(sess)
		}
	}
}

type styledText struct {
	text  string
	style lipgloss.Style
}

func render(r *lipgloss.Renderer, width int) string {
	artStyle := r.NewStyle().Foreground(lipgloss.Color("#666666"))
	nameStyle := r.NewStyle().Foreground(lipgloss.Color("#ffffff")).Bold(true)
	subStyle := r.NewStyle().Foreground(lipgloss.Color("#a0a0a0"))
	urlStyle := r.NewStyle().Foreground(lipgloss.Color("#87afaf"))
	dimStyle := r.NewStyle().Foreground(lipgloss.Color("#555555"))
	techStyle := r.NewStyle().Foreground(lipgloss.Color("#888888"))

	gap := "   "
	indent := "  "

	artLines := strings.Split(portrait, "\n")

	info := []styledText{
		{"", subStyle},
		{"Deniz Lopes Güneş", nameStyle},
		{"", subStyle},
		{"denizlg24.com", urlStyle},
		{"", subStyle},
		{"21 yo software engineer · porto, pt", subStyle},
		{"", subStyle},
		{"founder · ocean informatix", dimStyle},
		{"cs · feup · class of 2026", dimStyle},
		{"", subStyle},
		{"github.com/denizlg24", urlStyle},
		{"linkedin.com/in/deniz-güneş", urlStyle},
		{"denizlg24@gmail.com", urlStyle},
		{"", subStyle},
		{"next.js · typescript · react", techStyle},
		{"rust · postgresql · mongodb", techStyle},
	}

	var b strings.Builder
	b.WriteString("\n")

	for i, artLine := range artLines {
		line := indent + artStyle.Render(artLine)
		if i < len(info) && info[i].text != "" {
			line += gap + info[i].style.Render(info[i].text)
		}
		b.WriteString(line + "\n")
	}

	b.WriteString("\n")
	return b.String()
}
