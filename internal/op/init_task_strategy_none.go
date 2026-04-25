package op

import (
	// "fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// 未选择策略
func TelegramMessageHandlerStrategyNone(update tgbotapi.Update, bot *tgbotapi.BotAPI) error {
	// fmt.Printf("处理消息[none]: %s\n", update.Message.Text)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "监控面板未选择策略。")
	_, err := bot.Send(msg)
	return err
}
