package bot

import (
	"context"
	"errors"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"time-guard-bot/internal/helpers"
	"time-guard-bot/internal/models"
	"time-guard-bot/internal/storage/redis"
)

// Handles the /add command: /add {name} [desc]
func (b *Bot) HandleAddCommand(ctx context.Context, message *tgbotapi.Message, args []string) error {
	// Check if we have enough arguments
	if len(args) < 1 {
		return b.sendErrorMessage(message.Chat.ID, message.MessageID, "Please provide a name for the task")
	}

	// Check if the name is valid
	shortName := args[0]
	if err := helpers.ValidateShortName(shortName); err != nil {
		return b.sendErrorMessage(message.Chat.ID, message.MessageID, fmt.Sprintf("Invalid task name: %s", err.Error()))
	}

	// Extract description (all remaining arguments)
	description := ""
	if len(args) > 1 {
		description = strings.Join(args[1:], " ") // FIXME " "
	}

	// Check if we're at the limit of chat tasks
	count, err := b.storage.CountTasks(ctx, message.Chat.ID)
	if err != nil {
		return fmt.Errorf("failed to count tasks: %w", err)
	}

	if count >= helpers.MaxTasksPerChat {
		return b.sendErrorMessage(
			message.Chat.ID,
			message.MessageID,
			fmt.Sprintf("Maximum number of tasks per chat reached (%d)", helpers.MaxTasksPerChat),
		)
	}

	// Check if a task with this name already exists
	_, err = b.storage.GetTaskByName(ctx, message.Chat.ID, shortName)
	if err != nil {
		if errors.Is(err, redis.ErrNotFound) {
			// Continue
		} else {
			return fmt.Errorf("failed to check if task exists: %w", err)
		}
	} else {
		return b.sendErrorMessage(
			message.Chat.ID,
			message.MessageID,
			fmt.Sprintf("A task with name *%s* already exists", shortName),
		)
	}

	// Generate a unique ID for the task
	taskID, err := helpers.GenerateTaskID(helpers.TaskIDLength)
	if err != nil {
		return fmt.Errorf("failed to generate task ID: %w", err)
	}

	// Create task
	task := &models.Task{
		ID:          taskID,
		Name:        shortName,
		Description: description,
		ChatID:      message.Chat.ID,
		IsLocked:    false,
		LockReason:  "",
	}

	// Save task
	if err := b.storage.AddTask(ctx, task); err != nil {
		return fmt.Errorf("failed to add task: %w", err)
	}

	text := fmt.Sprintf("Task added successfully!\n\nName: *%s*\nID: `%s`", shortName, taskID)
	if description != "" {
		text += fmt.Sprintf("\nDescription: %s", description)
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ReplyToMessageID = message.MessageID
	msg.ParseMode = tgbotapi.ModeMarkdown
	_, err = b.api.Send(msg)

	return err
}

// Handles the /delete command: /delete {id}
func (b *Bot) HandleDeleteCommand(ctx context.Context, message *tgbotapi.Message, args []string) error {
	// Check if we have enough arguments
	if len(args) < 1 {
		return b.sendErrorMessage(message.Chat.ID, message.MessageID, "Please provide a task ID to delete")
	}

	taskID := args[0]

	// Get task to check if it exists
	task, err := b.storage.GetTask(ctx, message.Chat.ID, taskID)
	if err != nil {
		if errors.Is(err, redis.ErrNotFound) {
			return b.sendErrorMessage(message.Chat.ID, message.MessageID,
				fmt.Sprintf("Task with ID *%s* not found", taskID))
		}

		return fmt.Errorf("failed to get task: %w", err)
	}

	// Если task запущена
	if task.OwnerID != 0 {
		return b.sendErrorMessage(message.Chat.ID, message.MessageID,
			fmt.Sprintf("Task *%s* is currently in use", task.Name))
	}

	// Delete task
	if err := b.storage.DeleteTask(ctx, message.Chat.ID, taskID); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	text := fmt.Sprintf("Task deleted successfully!\n\nName: %s\nID: %s", task.Name, taskID)
	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ReplyToMessageID = message.MessageID
	_, err = b.api.Send(msg)

	return err
}

// Handles the /tasks command: /tasks
func (b *Bot) HandleTasksCommand(ctx context.Context, message *tgbotapi.Message, args []string) error {
	// Get all chat tasks
	tasks, err := b.storage.ListTasks(ctx, message.Chat.ID)
	if err != nil {
		return fmt.Errorf("failed to list tasks: %w", err)
	}

	if len(tasks) == 0 {
		msg := tgbotapi.NewMessage(message.Chat.ID, "No tasks found. Use /add to create a task")
		msg.ReplyToMessageID = message.MessageID
		_, err = b.api.Send(msg)

		return err
	}

	// Format tasks list
	text := "Tasks:\n\n"

	for _, task := range tasks {
		// Task status
		status := "🟢"
		if task.IsLocked {
			status = "🔒"
		} else if task.OwnerID != 0 {
			status = "⏱"
		}

		// Task info
		text += fmt.Sprintf("%s *%s* `%s` %s\n", status, task.Name, task.ID, task.Description)
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	_, err = b.api.Send(msg)

	return err
}
