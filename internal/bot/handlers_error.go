package bot

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// FIXME
// Sends an error message to a chat
func (b *Bot) sendErrorMessage(chatID int64, replyToID int, text string) error {
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("❌ %s", text))
	msg.ReplyToMessageID = replyToID
	msg.ParseMode = tgbotapi.ModeMarkdown
	_, err := b.api.Send(msg)
	return err
}
