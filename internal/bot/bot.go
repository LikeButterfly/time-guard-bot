// Copyright 2025 LikeButterfly
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package bot

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"time-guard-bot/internal/storage"
)

// Represents TimeGuardBot
type Bot struct {
	config   *Config
	api      *tgbotapi.BotAPI
	storage  storage.Storage
	handlers map[string]CommandHandler
	ctx      context.Context
	cancel   context.CancelFunc

	timers   map[string]*time.Timer
	timersMx sync.RWMutex

	wg sync.WaitGroup
}

// Represents bot configuration
type Config struct {
	Token           string
	LongPollTimeout int
}

// Represents a function that handles a bot command
type CommandHandler func(ctx context.Context, message *tgbotapi.Message, args []string) error

// Creates a new Bot instance
func NewBot(config *Config, storage storage.Storage) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot API: %w", err)
	}

	bot := &Bot{
		config:  config,
		api:     api,
		storage: storage,
		timers:  make(map[string]*time.Timer),
	}

	bot.registerHandlers()

	return bot, nil
}

// Starts the bot
func (b *Bot) Start() error {
	// Create context for bot
	ctx, cancel := context.WithCancel(context.Background())
	b.ctx = ctx
	b.cancel = cancel

	// Create update config
	updateConfig := tgbotapi.NewUpdate(-1)
	updateConfig.Timeout = b.config.LongPollTimeout

	updates := b.api.GetUpdatesChan(updateConfig)

	b.setBotCommands()

	// Restore active tasks
	log.Println("Restoring active timers...")

	if err := b.restoreActiveTasks(); err != nil {
		log.Printf("Error restoring timers: %v", err)
	} else {
		log.Println("Timers restored successfully")
	}

	// Start update loop in a goroutine
	go b.processUpdates(updates)

	return nil
}

// Stops the bot
func (b *Bot) Stop() {
	log.Println("Stopping bot gracefully...")

	if b.cancel != nil {
		b.cancel()
	}

	// Stopping all timers
	b.timersMx.Lock()
	for _, timer := range b.timers {
		timer.Stop()
	}
	b.timersMx.Unlock()

	// Waiting for the completion of all goroutines
	b.wg.Wait()
	log.Println("All goroutines stopped")
}

// Sets the bot commands in Telegram
func (b *Bot) setBotCommands() {
	commands := []tgbotapi.BotCommand{
		{Command: "add", Description: "Add a new task: /add name [description]"},
		{Command: "cancel", Description: "Cancel a task timer: /cancel [name]"},
		{Command: "status", Description: "Show task(s) status: /status [name]"},
		{Command: "tasks", Description: "List all tasks"},
		{Command: "lock", Description: "Lock a task: /lock id [reason]"},
		{Command: "unlock", Description: "Unlock a task: /unlock id"},
		{Command: "delete", Description: "Delete a task: /delete id"},
	}

	config := tgbotapi.NewSetMyCommands(commands...)

	_, err := b.api.Request(config)
	if err != nil {
		log.Printf("Failed to set bot commands: %v", err)
	}

	log.Printf("Set bot commands is successful")
}

// Processes updates from Telegram
func (b *Bot) processUpdates(updates tgbotapi.UpdatesChannel) {
	for {
		select {
		case <-b.ctx.Done():
			log.Println("Stopping updates processing...")
			return
		case update := <-updates:
			b.wg.Add(1)

			go func(u tgbotapi.Update) {
				defer b.wg.Done()

				b.processUpdate(b.ctx, u)
			}(update)
		}
	}
}

// Processes a single update from Telegram
func (b *Bot) processUpdate(ctx context.Context, update tgbotapi.Update) {
	// Process callback queries
	if update.CallbackQuery != nil {
		b.wg.Add(1)

		go func() {
			defer b.wg.Done()

			b.handleCallbackQuery(ctx, update.CallbackQuery)
		}()

		return
	}

	// Process messages
	if update.Message != nil {
		b.wg.Add(1)

		go func() {
			defer b.wg.Done()

			b.handleMessage(ctx, update.Message)
		}()

		return
	}
}

// Restores active tasks from storage
func (b *Bot) restoreActiveTasks() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	chats, err := b.storage.GetActiveChats(ctx)
	if err != nil {
		log.Printf("Failed to get chats with active tasks: %v", err)
	}

	if len(chats) == 0 {
		log.Printf("No active tasks found")
		return nil
	}

	log.Printf("Found %d chats with potential active tasks", len(chats))

	restoredCount := 0

	for _, chatID := range chats {
		// Get all active chat tasks
		activeTasks, err := b.storage.GetActiveTasks(ctx, chatID)
		if err != nil {
			log.Printf("Error retrieving active tasks: %v", err)
			continue
		}

		if len(activeTasks) == 0 {
			continue
		}

		// Restore each task's timer
		for _, task := range activeTasks {
			remaining := task.TimeRemaining()

			if remaining <= 0 {
				go b.handleTaskTimeout(ctx, task.ChatID, task.TaskID)
				continue
			}

			// Start a new timer with the remaining time
			b.startTaskTimer(ctx, task.ChatID, task.TaskID, time.Duration(remaining)*time.Second)

			restoredCount++
		}
	}

	log.Printf("Restored %d active timers", restoredCount)

	return nil
}
