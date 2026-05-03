package notify

import (
	"fmt"
	"strings"
	"time"

	"uptimepk/internal/db"
)

// 生成单个监控分组的汇总消息
func generateMonitorGroupSummary(monitorGid int64, domainName string) (string, error) {
	// 获取监控分组信息
	monitorGroup, err := db.GetMonitorGroupByID(monitorGid)
	if err != nil {
		return "", fmt.Errorf("failed to get monitor group: %v", err)
	}

	// 获取分组下的监控项
	monitors, err := db.GetMonitorListByGid(monitorGid)
	if err != nil {
		return "", fmt.Errorf("failed to get monitors: %v", err)
	}

	if len(monitors) == 0 {
		return "", nil
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
		logs, err := db.GetMonitorLogListByDate(monitor.ID, todayInt, 0, 0)

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
	message := fmt.Sprintf("分组: %s(%s)\n", monitorGroup.Name, groupURL)
	message += fmt.Sprintf("在线: %d, 离线: %d\n\n", onlineCount, offlineCount)
	message += "监控点详情:\n"
	message += monitorDetails
	message += "------------------------------------\n\n"

	return message, nil
}

// 生成单个监控分组的汇总消息
func GenerateMonitorGroupSummaryMessage(monitorGid int64) (string, error) {
	// 获取管理UI设置，用于获取域名
	domainName := ""
	settingAdminUI, err := db.GetSysSettingByCode(db.SettingAdminUI)
	if err == nil {
		adminUIValue, err := settingAdminUI.GetAdminUIValue()
		if err == nil && adminUIValue.DomainName != "" {
			domainName = strings.TrimRight(adminUIValue.DomainName, "/")
		}
	}

	// 构建消息内容
	message := fmt.Sprintf("📊 (%s): \n", time.Now().Format("2006-01-02 15:04:05"))
	message += "==========================\n\n"

	// 生成单个监控分组的汇总
	groupSummary, err := generateMonitorGroupSummary(monitorGid, domainName)
	if err != nil {
		return "", fmt.Errorf("failed to generate monitor group summary: %v", err)
	}

	if groupSummary == "" {
		return "", nil
	}

	message += groupSummary
	return message, nil
}

// 生成接收人监控汇总消息
func GenerateRecipientsSummaryMessage(recipientID int64) (string, error) {
	// 获取关联的监控分组
	relatedGroups, err := db.GetAdminRecipientsMonitorRelatedByRecipientID(recipientID)
	if err != nil {
		return "", fmt.Errorf("failed to get related monitor groups: %v", err)
	}

	if len(relatedGroups) == 0 {
		return "", nil
	}

	// 获取管理UI设置，用于获取域名
	domainName := ""
	settingAdminUI, err := db.GetSysSettingByCode(db.SettingAdminUI)
	if err == nil {
		adminUIValue, err := settingAdminUI.GetAdminUIValue()
		if err == nil && adminUIValue.DomainName != "" {
			domainName = strings.TrimRight(adminUIValue.DomainName, "/")
		}
	}

	// 构建消息内容
	message := fmt.Sprintf("📊 (%s): \n", time.Now().Format("2006-01-02 15:04:05"))
	message += "==========================\n\n"

	for _, related := range relatedGroups {
		groupSummary, err := generateMonitorGroupSummary(related.MonitorGid, domainName)
		if err != nil {
			continue
		}
		message += groupSummary
	}

	return message, nil
}
