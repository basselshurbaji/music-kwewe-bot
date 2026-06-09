// Command music-queue runs a Telegram bot that queues YouTube Music links and
// plays them in order via mpv.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"music-queue/internal/bot"
	"music-queue/internal/dotenv"
	"music-queue/internal/player"
	"music-queue/internal/queue"
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

	log.Println("music-queue is running; press Ctrl+C to stop")
	<-ctx.Done()
	log.Println("shutting down")
}
