package bot

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Registers all command handlers
func (b *Bot) registerHandlers() {
	b.handlers = map[string]CommandHandler{
		"start": b.handleStart,
	}
}

// Handles the /start command
func (b *Bot) handleStart(ctx context.Context, bot *Bot, message *tgbotapi.Message, args []string) error {
	welcomeText := "👋 Привет! Я TimeGuardBot - бот для управления временем задач\n\n"

	if message.Chat.IsGroup() || message.Chat.IsSuperGroup() {
		welcomeText += "Я помогу вашей группе управлять временем на задачи и избежать конфликтов\n\n"
	} else {
		welcomeText += "Я помогу вам управлять временем на задачи и быть более продуктивным\n\n"
	}

	welcomeText += "Используйте /help для получения списка доступных команд"

	// Send welcome message
	msg := tgbotapi.NewMessage(message.Chat.ID, welcomeText)
	_, err := b.api.Send(msg)
	return err
}
