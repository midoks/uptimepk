package op

import (
	"fmt"
	"strings"

	"uptimepk/internal/db"
	"uptimepk/internal/monitortask"
)

// InitRecipientsSummaryTasks 初始化接收人汇总任务
func InitRecipientsSummaryTasks() {
	// 获取所有启用的接收人
	recipients, _, err := db.GetAdminRecipientsList(1, 1000)
	if err != nil {
		fmt.Printf("Failed to get recipients list: %v\n", err)
		return
	}

	manager := monitortask.GetManager()

	for _, recipient := range recipients {
		if recipient.Status {
			// 生成cron表达式
			cronExpr := monitortask.GenerateCronExpr(recipient.Interval, recipient.IntervalType)

			// 创建任务
			task := monitortask.NewRecipientsSummaryTask(&recipient.AdminRecipients)

			// 添加到任务管理器
			if err := manager.AddTask(task, cronExpr); err != nil {
				fmt.Printf("failed to add recipients summary task for %d: %v\n", recipient.ID, err)
				SysLog(fmt.Sprintf("failed to add recipients summary task for %d: %v\n", recipient.ID, err))
			} else {
				SysLog(fmt.Sprintf("added recipients summary task for %d with cron: %s\n", recipient.ID, cronExpr))
			}
		}
	}
}

// ReloadRecipientsSummaryTasks 重新加载接收人汇总任务
func ReloadRecipientsSummaryTasks() {
	manager := monitortask.GetManager()

	// 移除所有现有的接收人汇总任务
	tasks := manager.ListTasks()
	for _, task := range tasks {
		if strings.HasPrefix(task.ID, "recipients_summary_") {
			if err := manager.RemoveTask(task.ID); err != nil {
				fmt.Printf("failed to remove recipients summary task %s: %v\n", task.ID, err)
			}
		}
	}

	// 重新添加所有任务
	InitRecipientsSummaryTasks()
}
