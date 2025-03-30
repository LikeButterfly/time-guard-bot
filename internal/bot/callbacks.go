package bot

import (
	"context"
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Sends an alert for a callback query
func (b *Bot) sendCallbackAlert(query *tgbotapi.CallbackQuery, text string) {
	callback := tgbotapi.NewCallback(query.ID, text)
	if _, err := b.api.Request(callback); err != nil {
		log.Printf("Failed to send callback alert: %v", err)
	}
}

// Handles callback queries from inline buttons
func (b *Bot) handleCallbackQuery(ctx context.Context, query *tgbotapi.CallbackQuery) {
	// Parse the callback data
	data := query.Data
	parts := strings.Split(data, ":")

	if len(parts) < 3 {
		log.Printf("Invalid callback data: %s", data)
		return
	}

	action := parts[0]

	switch action {
	case "check_time":
		taskID := parts[1]

		b.handleRemainingTimeCallback(ctx, query, taskID)
	default:
		log.Printf("Unknown callback action: %s", action)
	}
}

// Handles the check_time callback action
func (b *Bot) handleRemainingTimeCallback(ctx context.Context, query *tgbotapi.CallbackQuery, taskID string) {
	// Get active task
	activeTask, err := b.storage.GetActiveTask(ctx, query.Message.Chat.ID, taskID)
	if err != nil {
		b.sendCallbackAlert(query, "Task not found or not active")
		return
	}

	remaining := activeTask.TimeRemaining()

	// Calculate remaining time
	if remaining <= 0 {
		b.sendCallbackAlert(query, "Task time has expired")
		return
	}

	remainingMin := remaining / 60

	var remainingText string

	if remainingMin < 5 {
		remainingSec := remaining % 60
		remainingText = fmt.Sprintf("%d:%02d remaining", remainingMin, remainingSec)
	} else {
		remainingText = fmt.Sprintf("%d minutes remaining", remainingMin)
	}

	// Send alert with remaining time
	b.sendCallbackAlert(query, remainingText)
}
