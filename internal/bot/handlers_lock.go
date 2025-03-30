package bot

import (
	"context"
	"fmt"
	"strings"
	"time-guard-bot/internal/storage/redis"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Handles the /lock command: /lock {id} {reason}
func (b *Bot) HandleLockCommand(ctx context.Context, message *tgbotapi.Message, args []string) error {
	// Check if we have enough arguments
	if len(args) < 1 {
		return b.sendErrorMessage(message.Chat.ID, message.MessageID, "Please provide a task ID to lock")
	}

	taskID := args[0]

	// Optional reason
	reason := ""
	if len(args) > 1 {
		reason = strings.Join(args[1:], " ")
	}

	// Get task to lock
	task, err := b.storage.GetTask(ctx, message.Chat.ID, taskID)
	if err != nil {
		if err == redis.ErrNotFound {
			return b.sendErrorMessage(
				message.Chat.ID,
				message.MessageID,
				fmt.Sprintf("Task with ID '%s' not found", taskID),
			)
		}
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Check if already locked
	if task.IsLocked {
		return b.sendErrorMessage(
			message.Chat.ID,
			message.MessageID,
			fmt.Sprintf("Task *%s* is already locked", task.Name),
		)
	}

	if task.OwnerID != 0 {
		return b.sendErrorMessage(
			message.Chat.ID,
			message.MessageID,
			fmt.Sprintf("Task *%s* is currently in use", task.Name),
		)
	}

	// Lock task
	task.IsLocked = true
	task.LockReason = reason

	if err := b.storage.UpdateTask(ctx, task); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	// Send success message
	text := fmt.Sprintf("🔒 Task *%s* locked successfully", task.Name)
	if reason != "" {
		text += fmt.Sprintf("\nReason: %s", reason)
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ReplyToMessageID = message.MessageID
	_, err = b.api.Send(msg)
	return err
}

// Handles the /unlock command: /unlock {id}
func (b *Bot) HandleUnlockCommand(ctx context.Context, message *tgbotapi.Message, args []string) error {
	// Check if we have enough arguments
	if len(args) < 1 {
		return b.sendErrorMessage(message.Chat.ID, message.MessageID, "Please provide a task ID to unlock")
	}

	taskID := args[0]

	// Get task to unlock
	task, err := b.storage.GetTask(ctx, message.Chat.ID, taskID)
	if err != nil {
		if err == redis.ErrNotFound {
			return b.sendErrorMessage(message.Chat.ID, message.MessageID,
				fmt.Sprintf("Task with ID '%s' not found", taskID))
		}
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Check if already unlocked
	if !task.IsLocked {
		return b.sendErrorMessage(message.Chat.ID, message.MessageID,
			fmt.Sprintf("Task '%s' is not locked", task.Name))
	}

	// Unlock task
	task.IsLocked = false
	task.LockReason = ""

	if err := b.storage.UpdateTask(ctx, task); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	// Send success message
	text := fmt.Sprintf("🟢 Task '%s' unlocked successfully", task.Name)

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ReplyToMessageID = message.MessageID
	_, err = b.api.Send(msg)
	return err
}
