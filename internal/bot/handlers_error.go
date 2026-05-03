// Copyright 2025 LikeButterfly
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package bot

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Sends an error message to a chat
func (b *Bot) sendErrorMessage(chatID int64, replyToID int, text string) error {
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("❌ %s", text))
	msg.ReplyToMessageID = replyToID
	msg.ParseMode = tgbotapi.ModeMarkdown
	_, err := b.api.Send(msg)

	return err
}
