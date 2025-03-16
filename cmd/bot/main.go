package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"time-guard-bot/internal/bot"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Can't load .env file: %v", err)
	}

	tgToken := os.Getenv("TELEGRAM_TOKEN")
	if tgToken == "" {
		log.Fatal("TELEGRAM_TOKEN is required")
	}

	botConfig := &bot.Config{
		Token:           tgToken,
		LongPollTimeout: 60,
	}

	// Create bot
	b, err := bot.NewBot(botConfig)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	// Start bot
	log.Println("Starting bot...")
	if err := b.Start(); err != nil {
		log.Fatalf("Failed to start bot: %v", err)
	}

	// Create channel for receiving signals from the operating system
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// Blocking execution until a signal is received
	sig := <-sigCh
	log.Printf("Received signal %v, shutting down...", sig)

	log.Println("Bot stopped")
}
