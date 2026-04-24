package op

import (
	"fmt"
	"regexp"
	"strings"
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
func TelegramMessageHandlertrategyDefault(relateMonitorGroupID int64) tgtask.MessageHandler {
	return func(update tgbotapi.Update, bot *tgbotapi.BotAPI) error {
		fmt.Printf("处理消息[default] (groupID: %d): %s\n", relateMonitorGroupID, update.Message.Text)

		// 示例：根据消息内容做不同处理
		switch update.Message.Text {
		case "/start":
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "OK")
			_, err := bot.Send(msg)
			return err
		case "/status":
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "正常运行!")
			_, err := bot.Send(msg)
			return err
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
			// 尝试解析域名数据
			successCount, failCount, err := CreateMonitorsFromText(update.Message.Text, relateMonitorGroupID)
			if err != nil || successCount == 0 {
				if failCount == 0 {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "接收到数据: "+update.Message.Text)
					_, err := bot.Send(msg)
					return err
				}
			}

			// 返回创建结果
			var resultMsg string
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

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, resultMsg)
			_, err = bot.Send(msg)
			return err
		}
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
					handler = TelegramMessageHandlertrategyDefault(tp.RelateMonitorGroupID)
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
				handler = TelegramMessageHandlertrategyDefault(tp.RelateMonitorGroupID)
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

type DomainEntry struct {
	Remark string
	URL    string
}

// ParseDomainEntries 解析域名数据
func ParseDomainEntries(text string) ([]DomainEntry, error) {
	var entries []DomainEntry
	var currentRemark string

	lines := strings.Split(text, "\n")
	urlPattern := regexp.MustCompile(`^https?://`)

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		if strings.Contains(line, "=========================") {
			currentRemark = ""
			continue
		}

		// 检查这一行是否是 URL（以 http:// 或 https:// 开头）
		// 先移除反引号
		cleanLine := strings.Trim(line, "`")
		if urlPattern.MatchString(cleanLine) {
			if currentRemark != "" {
				entries = append(entries, DomainEntry{
					Remark: currentRemark,
					URL:    cleanLine,
				})
				currentRemark = ""
			}
		} else if strings.Contains(line, ":") || strings.Contains(line, "：") {
			// 这一行可能是备注行（支持英文和中文冒号）
			// 找到第一个冒号（英文或中文）
			colonIndex := strings.Index(line, ":")
			if colonIndex == -1 {
				colonIndex = strings.Index(line, "：")
			}
			if colonIndex != -1 {
				remark := strings.TrimSpace(line[:colonIndex])
				rest := strings.TrimSpace(line[colonIndex+1:])

				// 移除反引号
				rest = strings.Trim(rest, "`")

				if urlPattern.MatchString(rest) {
					// 备注和 URL 在同一行
					entries = append(entries, DomainEntry{
						Remark: remark,
						URL:    rest,
					})
					currentRemark = ""
				} else {
					// 这是备注行，URL 在下一行
					currentRemark = remark
				}
			}
		}
	}

	return entries, nil
}

func CreateMonitorsFromText(text string, gid int64) (successCount, failCount int, err error) {
	entries, err := ParseDomainEntries(text)

	if err != nil {
		return 0, 0, err
	}

	if len(entries) == 0 {
		return 0, 0, nil
	}

	for _, entry := range entries {
		monitor := &model.Monitor{
			Name:         entry.Remark,
			Type:         "http",
			Status:       true,
			Interval:     60,
			IntervalType: "second",
			MaxRetries:   3,
			Timeout:      10,
			Gid:          gid, // 添加关联 ID
			CreateTime:   time.Now().Unix(),
			UpdateTime:   time.Now().Unix(),
		}

		monitor.SetHttpTypeParams(model.MonitorHttpTypeParams{
			Addr: entry.URL,
		})

		if err := db.GetDb().Create(monitor).Error; err != nil {
			fmt.Printf("创建监控失败: %s - %v\n", entry.Remark, err)
			failCount++
			continue
		}

		if err := MonitorAddTask(*monitor); err != nil {
			fmt.Printf("添加任务失败: %s - %v\n", entry.Remark, err)
			failCount++
			continue
		}

		successCount++
	}

	return successCount, failCount, nil
}
