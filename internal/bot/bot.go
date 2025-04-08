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
	// Управление таймерами
	timers   map[string]*time.Timer
	timersMx sync.RWMutex
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
	updateConfig := tgbotapi.NewUpdate(0)
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
	if b.cancel != nil {
		b.cancel()
	}
}

// Sets the bot commands in Telegram
func (b *Bot) setBotCommands() {
	// Специально не добавляю start и help, чтобы было не так много команд (будет еще больше)
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
	for update := range updates {
		go b.processUpdate(b.ctx, update)
	}
}

// Processes a single update from Telegram
func (b *Bot) processUpdate(ctx context.Context, update tgbotapi.Update) {
	// Process callback queries
	if update.CallbackQuery != nil {
		go b.handleCallbackQuery(ctx, update.CallbackQuery)
		return
	}

	// Process messages
	if update.Message != nil {
		go b.handleMessage(ctx, update.Message)
		return
	}
}

// Restores active tasks from storage
func (b *Bot) restoreActiveTasks() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	groups, err := b.storage.GetActiveGroups(ctx)
	if err != nil {
		log.Printf("Failed to get groups with active tasks: %v", err)
	}

	if len(groups) == 0 {
		log.Printf("No active tasks found")
		return nil
	}

	log.Printf("Found %d groups with potential active tasks", len(groups))

	restoredCount := 0
	for _, groupID := range groups {
		// Get all active tasks for this group
		activeTasks, err := b.storage.GetActiveTasks(ctx, groupID)
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
				go b.handleTaskTimeout(task.GroupID, task.TaskID)
				continue
			}

			// Start a new timer with the remaining time
			b.startTaskTimer(task.GroupID, task.TaskID, time.Duration(remaining)*time.Second)
			restoredCount++
		}
	}

	log.Printf("Restored %d active timers", restoredCount)
	return nil
}
