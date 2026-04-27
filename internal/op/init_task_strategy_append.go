package op

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"uptimepk/internal/tgtask"
)

// 默认策略
func TelegramMessageHandlertrategyAppend(relateMonitorGroupID int64) tgtask.MessageHandler {
	return func(update tgbotapi.Update, bot *tgbotapi.BotAPI) error {
		// fmt.Printf("处理消息[default] (groupID: %d): %s\n", relateMonitorGroupID, update.Message.Text)

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
			return HandleHelpCommand(update, bot, true)
		default:
			var resultMsg string
			if relateMonitorGroupID == 0 {
				resultMsg = "未绑定监控分组,无法导入!"
			} else {
				// 尝试解析域名数据
				successCount, failCount, err := CreateMonitorsFromTextAppend(update.Message.Text, relateMonitorGroupID)
				if err != nil || successCount == 0 {
					if failCount == 0 {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "接收到数据: "+update.Message.Text)
						_, err := bot.Send(msg)
						return err
					}
				}

				if successCount > 0 {
					resultMsg = fmt.Sprintf("✓ 成功创建 %d 个监控任务", successCount)
				}
				if failCount > 0 {
					if resultMsg != "" {
						resultMsg += fmt.Sprintf("\n✗ 失败 %d 个", failCount)
					} else {
						resultMsg = fmt.Sprintf("✗ 失败 %d 个", failCount)
					}
				}
			}

			var sendErr error
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, resultMsg)
			_, sendErr = bot.Send(msg)
			return sendErr
		}
	}
}
