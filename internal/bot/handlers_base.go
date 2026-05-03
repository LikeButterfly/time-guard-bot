package bot

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"time-guard-bot/internal/helpers"
)

// Handles the /start command
func (b *Bot) HandleStartCommand(ctx context.Context, message *tgbotapi.Message, args []string) error {
	text := "👋 Hello! I'm TimeGuardBot - a bot for managing task time\n\n"

	text += "This bot helps your team track and manage time spent on various tasks\n\n"

	text += "Use /help to get a list of available commands"

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	_, err := b.api.Send(msg)

	return err
}

// Handles the /help command
func (b *Bot) HandleHelpCommand(ctx context.Context, message *tgbotapi.Message, args []string) error {
	text := ""

	text += "<b>Task Management</b>:\n"
	text += "/add {task_name} [task_desc] - Add a new task with a name and optional description\n"
	text += "/delete {task_id} - Delete a task by ID\n"
	text += "/tasks - List all tasks\n"
	text += "/status [task_name] - Show status of all tasks or a specific task\n"
	text += "/lock {task_id} [reason] - Lock a task, preventing it from being started\n"
	text += "/unlock {task_id} - Unlock a previously locked task\n\n"

	text += "<b>Time Tracking</b>:\n"
	text += "/{minutes} {task_name} - Start a timer for a task (e.g., '/30 coding')\n"
	text += "/cancel [task_name] - Cancel specified timer (defaults to latest)\n\n"

	text += "<b>Limits</b>:\n"
	text += fmt.Sprintf("- Maximum task duration: %d minutes (%.1f hours)\n", helpers.MaxTaskDuration, float64(helpers.MaxTaskDuration)/60)
	text += fmt.Sprintf("- Maximum active tasks per user: %d\n", helpers.MaxTasksPerUser)
	text += fmt.Sprintf("- Maximum tasks: %d", helpers.MaxTasksPerChat)

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ParseMode = tgbotapi.ModeHTML
	_, err := b.api.Send(msg)

	return err
}
