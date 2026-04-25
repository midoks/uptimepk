package op

import (
	"fmt"
	"time"

	"uptimepk/internal/db"
)

// InitCleanTask 初始化清理任务
func InitCleanTask() {
	// 每天凌晨 0 点执行清理
	go func() {
		for {
			// 计算到下一次凌晨 0 点的时间
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
			duration := next.Sub(now)
			// 等待到凌晨 0 点
			time.Sleep(duration)

			// time.Sleep(10 * time.Second)
			// 执行清理
			if err := CleanExpiredMonitorLogs(); err != nil {
				SysLog(fmt.Sprintf("[%s] 清理过期监控日志失败: %v", time.Now().Format("2006-01-02 15:04:05"), err))
			}
		}
	}()
}

// CleanExpiredMonitorLogs 清理过期的监控日志
func CleanExpiredMonitorLogs() error {
	// 获取数据库配置
	setting, err := db.GetSysSettingByCode(db.SettingDbConf)
	if err != nil {
		return fmt.Errorf("获取数据库配置失败: %v", err)
	}

	// 解析配置
	dbConf, err := setting.GetDbConfValue()
	if err != nil {
		return fmt.Errorf("解析数据库配置失败: %v", err)
	}

	// 获取天数（默认 180 天）
	days := dbConf.MonitorLogDays
	if days <= 0 {
		days = 180 // 默认 180 天
	}

	// 执行清理
	fmt.Printf("[%s] 开始清理 %d 天之前的监控日志\n", time.Now().Format("2006-01-02 15:04:05"), days)
	if err := db.DeleteMonitorLogBeforeDays(int(days)); err != nil {
		return fmt.Errorf("删除过期监控日志失败: %v", err)
	}

	fmt.Printf("[%s] 清理 %d 天之前的监控日志完成\n", time.Now().Format("2006-01-02 15:04:05"), days)
	return nil
}
