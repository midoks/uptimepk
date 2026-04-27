package monitortask

import (
	"context"
	"fmt"
	"strconv"

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
	return fmt.Sprintf("recipients Summary Task for %d", t.recipient.ID)
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

	// 生成消息内容
	message, err := notify.GenerateRecipientsSummaryMessage(t.recipient)
	if err != nil {
		return fmt.Errorf("failed to generate summary message: %v", err)
	}

	if message == "" {
		return nil
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
		return fmt.Sprintf("0 */%d * * * *", interval)
	case "hour":
		return fmt.Sprintf("0 0 */%d * * *", interval)
	case "day":
		return fmt.Sprintf("0 0 0 */%d * *", interval)
	default:
		return "0 */30 * * * *" // 默认每30分钟执行
	}
}
