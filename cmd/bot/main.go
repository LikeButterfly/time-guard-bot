// Copyright 2025 LikeButterfly
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/joho/godotenv"

	_ "time-guard-bot/docs/swagger"
	"time-guard-bot/internal/api"
	"time-guard-bot/internal/bot"
	"time-guard-bot/internal/storage"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using system environment variables: %v", err)
	}

	tgToken := os.Getenv("TELEGRAM_TOKEN")
	if tgToken == "" {
		log.Fatal("TELEGRAM_TOKEN is required")
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")

	redisDBStr := os.Getenv("REDIS_DB")
	redisDB := 0

	if redisDBStr != "" {
		var err error

		redisDB, err = strconv.Atoi(redisDBStr)
		if err != nil {
			log.Printf("Warning: invalid REDIS_DB, using default: %v", err)
		}
	}

	apiAddr := os.Getenv("API_ADDR")
	if apiAddr == "" {
		apiAddr = "0.0.0.0:8080"
	}

	// Create Redis storage
	redisStorage, err := storage.NewRedisStorage(redisAddr, redisPassword, redisDB)
	if err != nil {
		log.Fatalf("Failed to create Redis storage: %v", err)
	}
	defer redisStorage.Close()

	botConfig := &bot.Config{
		Token:           tgToken,
		LongPollTimeout: 60,
	}

	// Create bot
	b, err := bot.NewBot(botConfig, redisStorage)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	// Create API server
	apiConfig := &api.Config{
		Addr: apiAddr,
	}
	apiServer := api.NewServer(apiConfig, redisStorage)

	// Start API server
	if err := apiServer.Start(); err != nil {
		log.Fatalf("Failed to start API server: %v", err)
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

	// Stop API server
	if err := apiServer.Stop(); err != nil {
		log.Printf("Error stopping API server: %v", err)
	}

	// Stop bot
	b.Stop()

	log.Println("Bot stopped")
}
