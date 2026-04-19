package op

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"uptimepk/internal/db"
	"uptimepk/internal/tgtask"
)

// 未选择策略
func TelegramMessageHandlerStrategyNone(update tgbotapi.Update, bot *tgbotapi.BotAPI) error {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "未选择")
	_, err := bot.Send(msg)
	return err
}

// 默认策略
func TelegramMessageHandlertrategyDefault(update tgbotapi.Update, bot *tgbotapi.BotAPI) error {
	// 这里可以添加自定义的业务逻辑
	fmt.Printf("处理消息: %s\n", update.Message.Text)

	// 示例：根据消息内容做不同处理
	switch update.Message.Text {
	case "/start":
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Hello! I'm your uptime monitoring bot.")
		_, err := bot.Send(msg)
		return err
	case "/status":
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "All systems are running smoothly!")
		_, err := bot.Send(msg)
		return err
	default:
		// 默认回复
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "I received your message: "+update.Message.Text)
		_, err := bot.Send(msg)
		return err
	}
}

func InitTgTask() {
	manager := tgtask.GetManager()

	recipient_data, err := db.GetAdminRecipientsInstancesByID(1)
	if err != nil {
		fmt.Printf("Failed to get recipient data: %v\n", err)
		return
	}

	if recipient_data == nil {
		fmt.Println("No recipient data found")
		return
	}

	fmt.Printf("Recipient data: ID=%d, MediaType=%s\n", recipient_data.ID, recipient_data.MediaType)

	if recipient_data.MediaType == "telegram" {
		tp, err := recipient_data.GetTelegramParams()
		if err != nil {
			fmt.Printf("Failed to get telegram params: %v\n", err)
			return
		}

		// fmt.Printf("Telegram params: Token=%s, Proxy=%s\n", tp.Token, recipient_data.GetTelegramProxy())
		botID := recipient_data.ID
		if err := manager.AddBot(botID, tp.Token, recipient_data.GetTelegramProxy(), 0, TelegramMessageHandlertrategyDefault); err != nil {
			fmt.Printf("Failed to add bot: %v\n", err)
			return
		}

		if err := manager.StartBot(botID); err != nil {
			fmt.Printf("Failed to start bot: %v\n", err)
			return
		}

		fmt.Printf("Bot %d started successfully\n", botID)
	} else {
		fmt.Printf("MediaType is not telegram: %s\n", recipient_data.MediaType)
	}
}
