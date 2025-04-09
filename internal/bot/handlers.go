package bot

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"time-guard-bot/internal/helpers"
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
		"cancel": b.HandleCancelCommand,
	}
}

// Process a message from Telegram
func (b *Bot) handleMessage(ctx context.Context, message *tgbotapi.Message) {
	// Check if it's a command message
	if message.IsCommand() {
		command := message.Command()

		// Check if the command is a number (timer command)
		duration, err := strconv.Atoi(command)
		if err != nil {
			// Not a number, handle as a regular command
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

		// TODO обработать команды типа /1.01 {name}. Могут задудосить запросами в базу на поиск task с name ".01"

		// This is a timer command (e.g., /30 task_name)
		// Check that duration is positive and doesn't exceed the maximum limit
		if duration < helpers.MinTaskDuration {
			if err := b.sendErrorMessage(message.Chat.ID, message.MessageID, "Duration must be more than 1"); err != nil {
				log.Printf("Failed to send error message: %v", err)
			}

			return
		}

		if duration > helpers.MaxTaskDuration {
			if err := b.sendErrorMessage(
				message.Chat.ID,
				message.MessageID,
				fmt.Sprintf("Duration exceeds maximum allowed limit (%d minutes)", helpers.MaxTaskDuration),
			); err != nil {
				log.Printf("Failed to send error message: %v", err)
			}

			return
		}

		args := strings.Fields(message.CommandArguments())
		if len(args) == 0 {
			if err := b.sendErrorMessage(message.Chat.ID, message.MessageID, "Please provide a task name"); err != nil {
				log.Printf("Failed to send error message: %v", err)
			}

			return
		}

		if err := b.handleTimeCommand(ctx, message, duration, args[0]); err != nil {
			log.Printf("Error handling command %s: %v", command, err)

			if err := b.sendErrorMessage(message.Chat.ID, message.MessageID, "Error processing timer command"); err != nil {
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
