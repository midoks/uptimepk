package monitortask

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"uptimepk/internal/db"
	"uptimepk/internal/model"
	"uptimepk/internal/notify"
)

// RecipientsSummaryTask 监控汇总消息任务
type RecipientsSummaryTask struct {
	recipient *model.AdminRecipients
}

// NewRecipientsSummaryTask 创建新的监控汇总消息任务
func NewRecipientsSummaryTask(recipient *model.AdminRecipients) *RecipientsSummaryTask {
	return &RecipientsSummaryTask{
		recipient: recipient,
	}
}

// ID 返回任务ID
func (t *RecipientsSummaryTask) ID() string {
	return fmt.Sprintf("recipients_summary_%d", t.recipient.ID)
}

// Name 返回任务名称
func (t *RecipientsSummaryTask) Name() string {
	return fmt.Sprintf("Recipients Summary Task for %d", t.recipient.ID)
}

// Run 执行任务
func (t *RecipientsSummaryTask) Run() error {
	// 检查接收人状态
	if !t.recipient.Status {
		return nil
	}

	// 获取关联的监控分组
	relatedGroups, err := db.GetAdminRecipientsMonitorRelatedByRecipientID(t.recipient.ID)
	if err != nil {
		return fmt.Errorf("failed to get related monitor groups: %v", err)
	}

	if len(relatedGroups) == 0 {
		return nil
	}

	// 获取媒介实例信息
	mediaInstance, err := db.GetAdminRecipientsInstancesByID(t.recipient.MediaID)
	if err != nil {
		return fmt.Errorf("failed to get media instance: %v", err)
	}

	// 获取管理UI设置，用于获取域名
	domainName := ""
	settingAdminUI, err := db.GetSysSettingByCode("setting_admin_ui")
	if err == nil {
		adminUIValue, err := settingAdminUI.GetAdminUIValue()
		if err == nil && adminUIValue.DomainName != "" {
			domainName = strings.TrimRight(adminUIValue.DomainName, "/")
		}
	}

	// 构建消息内容
	message := fmt.Sprintf("📊 监控汇总报告\n时间: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

	for _, related := range relatedGroups {
		// 获取监控分组信息
		monitorGroup, err := db.GetMonitorGroupByID(related.MonitorGid)
		if err != nil {
			continue
		}

		// 获取分组下的监控项
		monitors, err := db.GetMonitorListByGid(related.MonitorGid)
		if err != nil {
			continue
		}

		if len(monitors) == 0 {
			continue
		}

		// 计算状态和可用率
		onlineCount := 0
		offlineCount := 0
		monitorDetails := ""

		for _, monitor := range monitors {
			if monitor.Status == 1 {
				onlineCount++
			} else {
				offlineCount++
			}

			// 获取最新监控日志
			latestLog, err := db.GetMonitorLatestLog(monitor.ID)
			currentStatus := "离线"
			if err == nil && latestLog != nil && latestLog.IsValid {
				currentStatus = "在线"
			}

			// 计算今天的可用率
			today := time.Now()
			year, month, day := today.Date()
			todayInt := int64(year*10000 + int(month)*100 + day)
			logs, err := db.GetMonitorLogListByDate(monitor.ID, todayInt)

			upRate := 0.0
			if err == nil && len(logs) > 0 {
				upCount := 0
				for _, log := range logs {
					if log.IsValid {
						upCount++
					}
				}
				upRate = float64(upCount) / float64(len(logs)) * 100
			}

			// 添加监控点详情
			monitorDetails += fmt.Sprintf("  - %s: %s (可用率: %.1f%%)\n", monitor.Name, currentStatus, upRate)
		}

		// 组装分组信息
		groupPath := fmt.Sprintf("/groups?id=%d", monitorGroup.ID)
		groupURL := groupPath
		if domainName != "" {
			groupURL = domainName + groupPath
		}
		message += fmt.Sprintf("📋 分组: %s\n", monitorGroup.Name)
		message += fmt.Sprintf("🌐 分组链接: %s\n", groupURL)
		message += fmt.Sprintf("📈 在线: %d, 离线: %d\n\n", onlineCount, offlineCount)
		message += "监控点详情:\n"
		message += monitorDetails
		message += "\n"
	}

	// 发送消息
	switch mediaInstance.MediaType {
	case "telegram":
		tp, err := mediaInstance.GetTelegramParams()
		if err != nil {
			return fmt.Errorf("failed to get telegram params: %v", err)
		}

		chatID, err := strconv.ParseInt(tp.SendID, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid chat ID: %v", err)
		}

		notification, err := notify.NewNotification(tp.Token, chatID, mediaInstance.GetTelegramProxy(), true)
		if err != nil {
			return fmt.Errorf("failed to create notification: %v", err)
		}

		if err := notification.Send(context.Background(), "监控汇总报告", message); err != nil {
			return fmt.Errorf("failed to send notification: %v", err)
		}
		// 可以扩展其他媒介类型
	}

	return nil
}

// GenerateCronExpr 根据接收人配置生成cron表达式
func GenerateCronExpr(interval int, intervalType string) string {
	switch intervalType {
	case "second":
		return fmt.Sprintf("*/%d * * * * *", interval)
	case "minute":
		return fmt.Sprintf("* */%d * * * *", interval)
	case "hour":
		return fmt.Sprintf("* * */%d * * *", interval)
	case "day":
		return fmt.Sprintf("* * * */%d * *", interval)
	default:
		return "* */30 * * * *" // 默认每秒执行
	}
}
