package op

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"uptimepk/internal/tgtask"
)

// 未选择策略
func TelegramMessageHandlerStrategyNone(relateMonitorGroupID int64) tgtask.MessageHandler {
	return func(update tgbotapi.Update, bot *tgbotapi.BotAPI) error {
		// 示例：根据消息内容做不同处理
		switch update.Message.Text {
		case "/status":
			return HandleStatusCommand(update, bot)
		case "/last":
			return HandleLastCommand(update, bot, relateMonitorGroupID)
		case "/start":
			fallthrough
		case "/?":
			fallthrough
		case "/help":
			return HandleHelpCommand(update, bot, false)
		default:
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "监控面板未选择策略。")
			_, err := bot.Send(msg)
			return err
		}
	}
}
