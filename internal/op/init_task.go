package op

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"uptimepk/internal/db"
	"uptimepk/internal/model"
	"uptimepk/internal/tgtask"
)

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
				case "append":
					handler = TelegramMessageHandlertrategyAppend(tp.RelateMonitorGroupID)
				default:
					handler = TelegramMessageHandlerStrategyNone(tp.RelateMonitorGroupID)
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
			case "append":
				handler = TelegramMessageHandlertrategyAppend(tp.RelateMonitorGroupID)
			default:
				handler = TelegramMessageHandlerStrategyNone(tp.RelateMonitorGroupID)
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

	// 清空之前监控任务
	gmonitor_list, err := db.GetMonitorListByGid(gid)
	if err == nil {
		for _, gm := range gmonitor_list {
			// fmt.Printf("软删除ID:%v\n", gm.ID)
			if err := db.MonitorSoftDeleteByID(gm.ID); err != nil { // 删除任务
				fmt.Printf("[Telegram]软删除失败:%v\n", err)
			}

			if err := MonitorDeleteTask(gm); err != nil {
				fmt.Printf("[Telegram]软删除任务失败:%v\n", err)
			}
		}
	}

	for _, entry := range entries {

		common_data := &model.Monitor{
			Name:         entry.Remark,
			Type:         "http",
			Status:       1,
			Interval:     60,
			IntervalType: "second",
			MaxRetries:   3,
			Timeout:      10,
			Gid:          gid, // 添加关联 ID
			Mark:         entry.Remark,
			CreateTime:   time.Now().Unix(),
			UpdateTime:   time.Now().Unix(),
		}

		common_data.SetHttpTypeParams(model.MonitorHttpTypeParams{
			Addr: entry.URL,
		})

		delete_id, err := db.GetMonitorDeletedID()

		if err == nil {
			if err := db.GetDb().Model(&model.Monitor{}).Where("id = ?", delete_id).Update("is_deleted", 0).Error; err != nil {
				continue
			}

			if err := db.GetDb().Where("id = ?", delete_id).Updates(common_data).Error; err != nil {
				fmt.Printf("创建监控失败: %s - %v\n", entry.Remark, err)
				failCount++
				continue
			}

		} else {
			if err := db.GetDb().Create(common_data).Error; err != nil {
				fmt.Printf("创建监控失败: %s - %v\n", entry.Remark, err)
				failCount++
				continue
			}
		}

		common_data.ID = delete_id

		// 删除之前的监控日志（必须在添加任务之前）
		if err := db.DeleteMonitorLogByMonitorID(common_data.ID); err != nil {
			fmt.Printf("删除过期监控日志失败: %s - %v\n", entry.Remark, err)
		}

		if err := MonitorAddTask(*common_data); err != nil {
			fmt.Printf("添加任务失败: %s - %v\n", entry.Remark, err)
			failCount++
			continue
		}
		successCount++
	}

	return successCount, failCount, nil
}

func CreateMonitorsFromTextAppend(text string, gid int64) (successCount, failCount int, err error) {
	entries, err := ParseDomainEntries(text)

	if err != nil {
		return 0, 0, err
	}

	if len(entries) == 0 {
		return 0, 0, nil
	}

	for _, entry := range entries {

		common_data := &model.Monitor{
			Name:         entry.Remark,
			Type:         "http",
			Status:       1,
			Interval:     60,
			IntervalType: "second",
			MaxRetries:   3,
			Timeout:      10,
			Gid:          gid, // 添加关联 ID
			Mark:         entry.Remark,
			CreateTime:   time.Now().Unix(),
			UpdateTime:   time.Now().Unix(),
		}

		common_data.SetHttpTypeParams(model.MonitorHttpTypeParams{
			Addr: entry.URL,
		})

		delete_id, err := db.GetMonitorDeletedID()

		if err == nil {
			if err := db.GetDb().Model(&model.Monitor{}).Where("id = ?", delete_id).Update("is_deleted", 0).Error; err != nil {
				continue
			}

			if err := db.GetDb().Where("id = ?", delete_id).Updates(common_data).Error; err != nil {
				fmt.Printf("创建监控失败: %s - %v\n", entry.Remark, err)
				failCount++
				continue
			}

		} else {
			if err := db.GetDb().Create(common_data).Error; err != nil {
				fmt.Printf("创建监控失败: %s - %v\n", entry.Remark, err)
				failCount++
				continue
			}
		}

		common_data.ID = delete_id

		// 删除之前的监控日志（必须在添加任务之前）
		if err := db.DeleteMonitorLogByMonitorID(common_data.ID); err != nil {
			fmt.Printf("删除过期监控日志失败: %s - %v\n", entry.Remark, err)
		}

		if err := MonitorAddTask(*common_data); err != nil {
			fmt.Printf("添加任务失败: %s - %v\n", entry.Remark, err)
			failCount++
			continue
		}
		successCount++
	}

	return successCount, failCount, nil
}
