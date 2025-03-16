package bot

import (
	"context"
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Represents TimeGuardBot
type Bot struct {
	config   *Config // FIXME возможно убрать отсюда
	api      *tgbotapi.BotAPI
	handlers map[string]CommandHandler
	ctx      context.Context
}

// Represents bot configuration
type Config struct {
	Token           string
	LongPollTimeout int
}

// Represents a function that handles a bot command
type CommandHandler func(ctx context.Context, b *Bot, message *tgbotapi.Message, args []string) error

// Creates a new Bot instance
func NewBot(config *Config) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot API: %w", err)
	}

	bot := &Bot{
		config: config,
		api:    api,
	}

	bot.registerHandlers()

	return bot, nil
}

// Starts the bot
func (b *Bot) Start() error {
	// Create context for bot
	b.ctx = context.Background()

	// Create update config
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = b.config.LongPollTimeout

	updates := b.api.GetUpdatesChan(updateConfig)

	b.setBotCommands()

	// b.restoreActiveTasks // TODO

	// Start update loop in a goroutine
	go b.processUpdates(updates)

	return nil
}

// Sets the bot commands in Telegram
func (b *Bot) setBotCommands() {
	commands := []tgbotapi.BotCommand{
		{Command: "help", Description: "Показать справку"},
	}

	config := tgbotapi.NewSetMyCommands(commands...)
	_, err := b.api.Request(config)
	if err != nil {
		log.Printf("Failed to set bot commands: %v", err)
	}

	log.Printf("Set bot commands is successful")

	// TODO подумать стоит ли fatal'ить
}

// Processes updates from Telegram
func (b *Bot) processUpdates(updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		go b.processUpdate(b.ctx, update)
	}
}

// Processes a single update from Telegram
func (b *Bot) processUpdate(ctx context.Context, update tgbotapi.Update) {
	// TODO // Process callback queries (button presses)
	// ...

	// Process messages
	if update.Message != nil {
		// Ignore non-command messages or edited messages
		if !update.Message.IsCommand() && update.Message.ReplyToMessage == nil {
			log.Printf("Ignore message: %v", update.Message)
			return
		}

		go b.handleMessage(ctx, update.Message)
		return
	}
}

// Processes a message
func (b *Bot) handleMessage(ctx context.Context, message *tgbotapi.Message) {
	// Check if it's a command message
	if message.IsCommand() {
		command := message.Command()
		args := strings.Fields(message.CommandArguments())

		// Find handler for this command
		handler, exists := b.handlers[command]
		if !exists {
			// Command not found, ignore it as per requirements
			log.Printf("Ignore not found command: %v", command) // TODO - delete debug log or do this log to debug
			return
		}

		// Execute handler
		if err := handler(ctx, b, message, args); err != nil {
			log.Printf("Error handling command %s: %v", command, err)
			b.sendErrorMessage(message.Chat.ID, message.MessageID, "Error processing command. Please try again")
		}
		return
	}

	// TODO // Check if it's a reply to a message
	// ...
}

// Sends an error message to a chat
func (b *Bot) sendErrorMessage(chatID int64, replyToID int, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if replyToID != 0 {
		msg.ReplyToMessageID = replyToID
	}

	_, err := b.api.Send(msg)
	if err != nil {
		log.Printf("Can't send message: %v", msg) // TODO обработать
	}
}
