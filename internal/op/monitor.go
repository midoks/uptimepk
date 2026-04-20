package op

import (
	"fmt"
	"strconv"

	"uptimepk/internal/db"
	"uptimepk/internal/model"
	"uptimepk/internal/monitortask"
)

// MonitorTask 监控任务
type MonitorTask struct {
	monitor *model.Monitor
}

// ID 获取任务ID
func (t *MonitorTask) ID() string {
	return "monitor_" + strconv.FormatInt(t.monitor.ID, 10)
}

// Name 获取任务名称
func (t *MonitorTask) Name() string {
	return t.monitor.Name
}

// Run 执行任务
func (t *MonitorTask) Run() error {
	switch t.monitor.Type {
	case "http":
		return t.runHttpMonitor()
	case "tcp":
		return t.runTcpMonitor()
	case "udp":
		return t.runUdpMonitor()
	default:
		return fmt.Errorf("unsupported monitor type: %s", t.monitor.Type)
	}
}

// runHttpMonitor 执行HTTP监控
func (t *MonitorTask) runHttpMonitor() error {
	params, err := t.monitor.GetHttpTypeParams()
	if err != nil {
		return err
	}
	// TODO: 实现HTTP监控逻辑
	fmt.Printf("Running HTTP monitor for %s: %s\n", t.monitor.Name, params.Addr)
	return nil
}

// runTcpMonitor 执行TCP监控
func (t *MonitorTask) runTcpMonitor() error {
	params, err := t.monitor.GetTcpTypeParams()
	if err != nil {
		return err
	}
	// TODO: 实现TCP监控逻辑
	fmt.Printf("Running TCP monitor for %s: %s:%d\n", t.monitor.Name, params.Host, params.Port)
	return nil
}

// runUdpMonitor 执行UDP监控
func (t *MonitorTask) runUdpMonitor() error {
	params, err := t.monitor.GetUdpTypeParams()
	if err != nil {
		return err
	}
	// TODO: 实现UDP监控逻辑
	fmt.Printf("Running UDP monitor for %s: %s:%d\n", t.monitor.Name, params.Host, params.Port)
	return nil
}

// InitMonitorask 初始化监控任务
func InitMonitorask() {
	manager := monitortask.GetManager()

	// 从数据库获取所有监控
	var monitors []model.Monitor
	if err := db.GetDb().Find(&monitors).Error; err != nil {
		fmt.Printf("Failed to get monitor list: %v\n", err)
		return
	}

	// 为每个监控创建任务
	for _, monitor := range monitors {
		fmt.Println("monitor:", monitor)
		if !monitor.Status {
			continue // 跳过禁用的监控
		}

		task := &MonitorTask{monitor: &monitor}

		// 根据监控间隔生成cron表达式
		// 例如：每60秒执行一次 -> "*/60 * * * * *"
		// 使用6字段cron表达式（秒、分、时、日、月、周）
		cronExpr := fmt.Sprintf("*/%d * * * * *", monitor.Interval)

		// 添加任务到管理器
		if err := manager.AddTask(task, cronExpr); err != nil {
			fmt.Printf("Failed to add monitor task %s: %v\n", monitor.Name, err)
			continue
		}

		fmt.Printf("Added monitor task %s with interval %d seconds\n", monitor.Name, monitor.Interval)
	}

	// 启动任务管理器
	manager.Start()
	fmt.Println("Monitor tasks initialized")
}
