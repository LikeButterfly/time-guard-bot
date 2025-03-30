package bot

import (
	"context"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Registers all command handlers
func (b *Bot) registerHandlers() {
	b.handlers = map[string]CommandHandler{
		"start":  b.HandleStartCommand,
		"help":   b.HandleHelpCommand,
		"add":    b.HandleAddCommand,
		"delete": b.HandleDeleteCommand,
		"tasks":  b.HandleTasksCommand,
		"status": b.HandleStatusCommand,
		"lock":   b.HandleLockCommand,
		"unlock": b.HandleUnlockCommand,
	}
}

// Process a message from Telegram
func (b *Bot) handleMessage(ctx context.Context, message *tgbotapi.Message) {
	// Check if it's a command message
	if message.IsCommand() {
		command := message.Command()

		// Check if the command is a number (timer command)
		duration, err := strconv.Atoi(command)
		// TODO еще проверить что duration не превышает лимит
		if err == nil && duration > 0 {
			// This is a timer command (e.g., /30 taskname)
			args := strings.Fields(message.CommandArguments())
			if len(args) > 0 {
				if err := b.handleTimeCommand(ctx, message, duration, args[0]); err != nil {
					// FIXME..
					if err := b.sendErrorMessage(message.Chat.ID, message.MessageID, "Error processing timer command"); err != nil {
						log.Printf("Failed to send error message: %v", err)
					}
				}
				return
			}
		}

		// Handle regular commands
		handler, exists := b.handlers[command]
		if !exists {
			// Command not found, ignore it as per requirements
			log.Printf("Ignore not found command: %v", command)
			return
		}

		// Execute handler
		args := strings.Fields(message.CommandArguments())
		if err := handler(ctx, message, args); err != nil {
			log.Printf("Error handling command %s: %v", command, err)
			if err := b.sendErrorMessage(message.Chat.ID, message.MessageID, "Error processing command. Please try again"); err != nil {
				log.Printf("Failed to send error message: %v", err)
			}
		}
		return
	}

	// Handle reply to message
	if message.ReplyToMessage != nil {
		b.handleReplyMessage(ctx, message)
		return
	}
}

// Handles a reply to a message
func (b *Bot) handleReplyMessage(ctx context.Context, message *tgbotapi.Message) {
	// TODO: Implement reply handling
	log.Printf("Ignoring reply message: %v", message)
}
