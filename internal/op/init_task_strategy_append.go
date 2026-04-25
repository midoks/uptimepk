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
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "正常运行!")
			_, err := bot.Send(msg)
			return err
		case "/start":
			fallthrough
		case "/?":
			fallthrough
		case "/help":
			helpText := `可用命令:
/start - 开始使用
/status - 检查运行状态
/help - 显示此帮助信息

批量导入格式:
备注: https://example.com
备注: https://test.com
=========================
备注: https://domain.com

说明:
- 每行格式: 备注: URL
- 使用 ========================= 作为分组分隔符
- URL 必须以 http:// 或 https:// 开头`
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, helpText)
			_, err := bot.Send(msg)
			return err
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
