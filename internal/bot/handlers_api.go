package bot

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"time-guard-bot/internal/helpers"
)

// Handles the /api_key command
func (b *Bot) HandleAPICommand(ctx context.Context, message *tgbotapi.Message, args []string) error {
	chatID := message.Chat.ID

	apiKey := helpers.GenerateAPIKey(chatID)

	responseMsg := fmt.Sprintf("`%s`", apiKey)

	msg := tgbotapi.NewMessage(message.Chat.ID, responseMsg)
	msg.ParseMode = "Markdown"
	_, err := b.api.Send(msg)

	return err
}
