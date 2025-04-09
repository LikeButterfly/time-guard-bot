package bot

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"time-guard-bot/internal/models"
	"time-guard-bot/internal/storage/redis"
)

// Handles the /status command: /status [name]
func (b *Bot) HandleStatusCommand(ctx context.Context, message *tgbotapi.Message, args []string) error {
	// If no arguments, get status of all tasks
	if len(args) == 0 {
		return b.getAllTasksStatus(ctx, message)
	}

	// Get status of specific task
	taskName := args[0]

	return b.getTaskStatus(ctx, message, taskName)
}

// Get status of a specific task
func (b *Bot) getTaskStatus(ctx context.Context, message *tgbotapi.Message, taskName string) error {
	// Try to get by short name
	task, err := b.storage.GetTaskByName(ctx, message.Chat.ID, taskName)
	if err != nil {
		if errors.Is(err, redis.ErrNotFound) {
			return b.sendErrorMessage(
				message.Chat.ID,
				message.MessageID,
				fmt.Sprintf("Task *%s* not found", taskName),
			)
		}

		return fmt.Errorf("failed to get task: %w", err)
	}

	var statusEmoji string

	var statusInfo string

	if task.IsLocked {
		statusEmoji = "🔒"

		statusInfo = "Locked"
		if task.LockReason != "" {
			statusInfo += fmt.Sprintf(" (%s)", task.LockReason)
		}
	} else {
		if task.OwnerID != 0 {
			statusEmoji = "⏱"
			// Get active task info for remaining time
			remaining := task.TimeRemaining()

			var remainingTime string

			if remaining < 60 {
				remainingTime = "< 1 min."
			} else {
				remainingTime = fmt.Sprintf("%d min.", remaining/60)
			}

			statusInfo = fmt.Sprintf("Remaining: %s", remainingTime)
		} else {
			statusEmoji = "🟢"
			statusInfo = "Available"
		}
	}

	text := fmt.Sprintf("%s %s\n", statusEmoji, statusInfo)

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ReplyToMessageID = message.MessageID
	_, err = b.api.Send(msg)

	return err
}

// Get status of all tasks in a group
func (b *Bot) getAllTasksStatus(ctx context.Context, message *tgbotapi.Message) error {
	tasks, err := b.storage.ListTasks(ctx, message.Chat.ID)
	if err != nil {
		return fmt.Errorf("failed to get tasks: %w", err)
	}

	if len(tasks) == 0 {
		return b.sendErrorMessage(
			message.Chat.ID,
			message.MessageID,
			"No tasks found for this chat",
		)
	}

	// Get active tasks
	activeTasks, err := b.storage.GetActiveTasks(ctx, message.Chat.ID)
	if err != nil {
		log.Printf("Failed to get active tasks: %v", err)

		activeTasks = []*models.ActiveTask{} // Empty slice as fallback
	}

	// Create a map for easier access to active tasks by ID
	activeTasksMap := make(map[string]*models.ActiveTask)
	for _, task := range activeTasks {
		activeTasksMap[task.TaskID] = task
	}

	// Build status message
	var text strings.Builder

	text.WriteString("Tasks Status:\n\n")

	for _, task := range tasks {
		// Get status emoji
		var statusEmoji string

		var statusInfo string

		if task.IsLocked {
			statusEmoji = "🔒"

			statusInfo = "Locked"
			if task.LockReason != "" {
				statusInfo += fmt.Sprintf(" (%s)", task.LockReason)
			}
		} else {
			if task.OwnerID != 0 {
				statusEmoji = "⏱"

				// Get active task info for remaining time
				activeTask, exists := activeTasksMap[task.ID]
				if exists {
					remaining := activeTask.TimeRemaining()

					var remainingTime string

					if remaining < 60 {
						remainingTime = "< 1 min."
					} else {
						remainingTime = fmt.Sprintf("%d min.", remaining/60)
					}

					statusInfo = fmt.Sprintf("Remaining: %s", remainingTime)
				} else {
					statusInfo = "Unexpected" // TODO обработать
				}
			} else {
				statusEmoji = "🟢"
				statusInfo = "Available"
			}
		}

		taskLine := fmt.Sprintf("%s *%s* - %s\n", statusEmoji, task.Name, statusInfo)
		text.WriteString(taskLine)
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, text.String())
	msg.ParseMode = tgbotapi.ModeMarkdown
	_, err = b.api.Send(msg)

	return err
}
