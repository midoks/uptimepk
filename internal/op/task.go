package op

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"uptimepk/internal/db"
	"uptimepk/internal/tgtask"
)

// 未选择策略
func TelegramMessageHandlerStrategyNone(update tgbotapi.Update, bot *tgbotapi.BotAPI) error {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "监控面板未选择策略。")
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

	telegram_list, err := db.GetAdminRecipientsInstancesByTelegram()
	if err != nil {
		fmt.Printf("failed to get recipient data: %v\n", err)
		return
	}

	if len(telegram_list) == 0 {
		return
	}

	for _, data := range telegram_list {
		tp, err := data.GetTelegramParams()
		if err != nil {
			fmt.Printf("failed to get telegram params: %v\n", err)
			continue
		}

		if tp.TelegramListenEnable {
			botID := data.ID

			if tp.TelegramListenStrategy == "" || tp.TelegramListenStrategy == "none" {
				if err := manager.AddBot(botID, tp.Token, data.GetTelegramProxy(), 0, TelegramMessageHandlerStrategyNone); err != nil {
					fmt.Printf("failed to add bot: %v\n", err)
					continue
				}
			}

			if tp.TelegramListenStrategy == "default" {
				if err := manager.AddBot(botID, tp.Token, data.GetTelegramProxy(), 0, TelegramMessageHandlertrategyDefault); err != nil {
					fmt.Printf("failed to add bot: %v\n", err)
					continue
				}
			}

			if err := manager.StartBot(botID); err != nil {
				fmt.Printf("failed to start bot: %v\n", err)
				continue
			}

			fmt.Printf("Bot %d started successfully\n", botID)
		}

	}

}
