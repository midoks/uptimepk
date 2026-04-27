package op

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"uptimepk/internal/db"
	"uptimepk/internal/notify"
)

// HandleStatusCommand 处理/status命令
func HandleStatusCommand(update tgbotapi.Update, bot *tgbotapi.BotAPI) error {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "正常运行!")
	_, err := bot.Send(msg)
	return err
}

// HandleLastCommand 处理/last命令
func HandleLastCommand(update tgbotapi.Update, bot *tgbotapi.BotAPI, relateMonitorGroupID int64) error {
	// 获取接收人信息
	var message string
	var err error
	if relateMonitorGroupID > 0 {
		// 先尝试通过监控分组ID获取接收人
		recipients, err := db.GetAdminRecipientsByMonitorGid(relateMonitorGroupID)
		if err == nil && len(recipients) > 0 {
			// 使用第一个接收人生成汇总消息
			message, err = notify.GenerateRecipientsSummaryMessage(recipients[0].ID)
		} else {
			// 如果没有关联的接收人，直接使用监控分组ID生成汇总消息
			message, err = notify.GenerateMonitorGroupSummaryMessage(relateMonitorGroupID)
		}
	}

	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "生成汇总消息失败: "+err.Error())
		_, err := bot.Send(msg)
		return err
	}

	if message == "" {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "暂无监控数据")
		_, err := bot.Send(msg)
		return err
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, message)
	_, err = bot.Send(msg)
	return err
}

// HandleHelpCommand 处理/help命令
func HandleHelpCommand(update tgbotapi.Update, bot *tgbotapi.BotAPI, includeImportFormat bool) error {
	helpText := `可用命令:
/start - 开始使用
/status - 检查运行状态
/last - 获取监控汇总信息
/help - 显示此帮助信息`

	if includeImportFormat {
		helpText += `

批量导入格式:
备注: https://example.com
备注: https://test.com
=========================
备注: https://domain.com

说明:
- 每行格式: 备注: URL
- 使用 ========================= 作为分组分隔符
- URL 必须以 http:// 或 https:// 开头`
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, helpText)
	_, err := bot.Send(msg)
	return err
}
