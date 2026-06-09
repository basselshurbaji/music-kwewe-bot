// Command music-queue runs a Telegram bot that queues YouTube Music links and
// plays them in order via mpv.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"music-queue/internal/bot"
	"music-queue/internal/dotenv"
	"music-queue/internal/player"
	"music-queue/internal/queue"
	"music-queue/internal/web"
)

func main() {
	if err := dotenv.Load(".env"); err != nil {
		log.Fatalf("loading .env: %v", err)
	}

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is not set")
	}

	q := queue.New()
	p := player.New(q)

	b, err := bot.New(token, q, p)
	if err != nil {
		log.Fatalf("init bot: %v", err)
	}
	log.Printf("authorized as @%s", b.Username())

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go p.Run(ctx)
	go b.Run()

	// Read-only dashboard.
	addr := os.Getenv("DASHBOARD_ADDR")
	if addr == "" {
		addr = ":7070"
	}
	srv := &http.Server{Addr: addr, Handler: web.New(q, p).Handler()}
	go func() {
		log.Printf("dashboard on http://localhost%s", port(addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("dashboard server: %v", err)
		}
	}()
	openBrowser("http://localhost" + port(addr))

	log.Println("music-kwewe-bot is running; press Ctrl+C to stop")
	<-ctx.Done()
	log.Println("shutting down")
}

// port normalizes an http addr (e.g. ":8080" or "0.0.0.0:8080") to a ":port"
// suffix suitable for building a localhost URL.
func port(addr string) string {
	if i := strings.LastIndex(addr, ":"); i >= 0 {
		return addr[i:]
	}
	return ":" + addr
}

// openBrowser best-effort opens url in the default browser.
func openBrowser(url string) {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	case "windows":
		cmd, args = "rundll32", []string{"url.dll,FileProtocolHandler"}
	default:
		cmd = "xdg-open"
	}
	if err := exec.Command(cmd, append(args, url)...).Start(); err != nil {
		log.Printf("could not open browser (%s): %v", url, err)
	}
}
