package op

import (
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"uptimepk/internal/db"
	"uptimepk/internal/model"
	"uptimepk/internal/tgtask"
)

// 未选择策略
func TelegramMessageHandlerStrategyNone(update tgbotapi.Update, bot *tgbotapi.BotAPI) error {
	fmt.Printf("处理消息[none]: %s\n", update.Message.Text)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "监控面板未选择策略。")
	_, err := bot.Send(msg)
	return err
}

// 默认策略
func TelegramMessageHandlertrategyDefault(update tgbotapi.Update, bot *tgbotapi.BotAPI) error {
	fmt.Printf("处理消息[default]: %s\n", update.Message.Text)

	// 示例：根据消息内容做不同处理
	switch update.Message.Text {
	case "/start":
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "YES.")
		_, err := bot.Send(msg)
		return err
	case "/status":
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "正常运行!")
		_, err := bot.Send(msg)
		return err
	default:
		// 默认回复
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "接收到数据: "+update.Message.Text)
		_, err := bot.Send(msg)
		return err
	}
}

func ReloadTelegramTask() {
	manager := tgtask.GetManager()

	// 获取当前的 Telegram 配置
	telegram_list, err := db.GetAdminRecipientsInstancesByTelegram()
	if err != nil {
		fmt.Printf("failed to get recipient data: %v\n", err)
		return
	}

	// 创建当前配置的映射，用于快速查找
	currentConfigs := make(map[int64]*model.AdminMediaInstance)
	for i := range telegram_list {
		currentConfigs[telegram_list[i].ID] = &telegram_list[i]
	}

	// 移除所有现有的机器人
	if err := manager.RemoveAllBots(); err != nil {
		fmt.Printf("failed to remove all bots: %v\n", err)
	}

	// 等待足够的时间确保所有 bot 实例完全停止
	time.Sleep(3 * time.Second)

	// 重新添加和启动机器人
	if len(telegram_list) > 0 {
		for _, data := range telegram_list {
			tp, err := data.GetTelegramParams()
			if err != nil {
				fmt.Printf("failed to get telegram params: %v\n", err)
				continue
			}

			if tp.TelegramListenEnable {
				botID := data.ID

				var handler tgtask.MessageHandler
				switch tp.TelegramListenStrategy {
				case "default":
					handler = TelegramMessageHandlertrategyDefault
				default:
					handler = TelegramMessageHandlerStrategyNone
				}

				if err := manager.AddBot(botID, tp.Token, data.GetTelegramProxy(), 0, handler); err != nil {
					fmt.Printf("failed to add bot: %v\n", err)
					continue
				}

				if err := manager.StartBot(botID); err != nil {
					fmt.Printf("failed to start bot: %v\n", err)
					continue
				}
				// fmt.Printf("Bot %d reloaded successfully\n", botID)
			}
		}
	}
}

func InitTelegramTask() {
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

			var handler tgtask.MessageHandler
			switch tp.TelegramListenStrategy {
			case "default":
				handler = TelegramMessageHandlertrategyDefault
			default:
				handler = TelegramMessageHandlerStrategyNone
			}

			if err := manager.AddBot(botID, tp.Token, data.GetTelegramProxy(), 0, handler); err != nil {
				fmt.Printf("failed to add bot: %v\n", err)
				continue
			}

			if err := manager.StartBot(botID); err != nil {
				fmt.Printf("failed to start bot: %v\n", err)
				continue
			}

			// fmt.Printf("Bot %d started successfully\n", botID)
		}

	}
}
