package op

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"time"

	"uptimepk/internal/db"
	"uptimepk/internal/model"
	"uptimepk/internal/monitortask"
	"uptimepk/internal/utils"
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
		// 插入错误监控日志
		if err := db.CreateMonitorLog(t.monitor.ID, false, "0", "0s", err.Error(), t.monitor.MaxRetries); err != nil {
			fmt.Printf("Failed to insert monitor log: %v\n", err)
		}

		return err
	}

	// 创建HTTP客户端，设置超时时间
	client := utils.NewHTTPClient(time.Duration(t.monitor.Timeout) * time.Second)

	// 记录开始时间
	startTime := time.Now()

	// 发送HTTP请求
	resp, err := client.Get(params.Addr)
	if err != nil {
		// 插入错误监控日志
		if err := db.CreateMonitorLog(t.monitor.ID, false, "0", time.Since(startTime).String(), err.Error(), t.monitor.MaxRetries); err != nil {
			fmt.Printf("Failed to insert monitor log: %v\n", err)
		}

		return fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	// 计算消耗时间
	duration := time.Since(startTime)

	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		// 插入错误监控日志
		if err := db.CreateMonitorLog(t.monitor.ID, false, "0", duration.String(), err.Error(), t.monitor.MaxRetries); err != nil {
			fmt.Printf("Failed to insert monitor log: %v\n", err)
		}

		return fmt.Errorf("failed to read response body: %v", err)
	}

	// 记录监控结果
	fmt.Printf("HTTP monitor for %s: %s\n", t.monitor.Name, params.Addr)
	fmt.Printf("Status code: %d\n", resp.StatusCode)
	fmt.Printf("Response time: %v\n", duration)
	fmt.Printf("Response size: %d bytes\n", len(body))

	// 检查状态码是否正常
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		fmt.Printf("HTTP monitor %s: OK\n", t.monitor.Name)
	} else {
		fmt.Printf("HTTP monitor %s: WARNING - Status code %d\n", t.monitor.Name, resp.StatusCode)
	}

	// 插入监控日志
	if err := db.CreateMonitorLog(t.monitor.ID, resp.StatusCode >= 200 && resp.StatusCode < 300, strconv.Itoa(len(body)), duration.String(), "", t.monitor.MaxRetries); err != nil {
		fmt.Printf("Failed to insert monitor log: %v\n", err)
	}

	return nil
}

// runTcpMonitor 执行TCP监控
func (t *MonitorTask) runTcpMonitor() error {
	params, err := t.monitor.GetTcpTypeParams()
	if err != nil {
		// 插入错误监控日志
		if err := db.CreateMonitorLog(t.monitor.ID, false, "0", "0s", err.Error(), t.monitor.MaxRetries); err != nil {
			fmt.Printf("Failed to insert monitor log: %v\n", err)
		}

		return err
	}

	// 记录开始时间
	startTime := time.Now()

	// 连接TCP服务器
	addr := fmt.Sprintf("%s:%d", params.Host, params.Port)
	conn, err := net.DialTimeout("tcp", addr, time.Duration(t.monitor.Timeout)*time.Second)
	if err != nil {
		// 插入错误监控日志
		if err := db.CreateMonitorLog(t.monitor.ID, false, "0", time.Since(startTime).String(), err.Error(), t.monitor.MaxRetries); err != nil {
			fmt.Printf("Failed to insert monitor log: %v\n", err)
		}

		return fmt.Errorf("TCP connection failed: %v", err)
	}
	defer conn.Close()

	// 计算消耗时间
	duration := time.Since(startTime)

	// 记录监控结果
	fmt.Printf("TCP monitor for %s: %s:%d\n", t.monitor.Name, params.Host, params.Port)
	fmt.Printf("Response time: %v\n", duration)
	fmt.Printf("TCP monitor %s: OK\n", t.monitor.Name)

	// 插入监控日志
	if err := db.CreateMonitorLog(t.monitor.ID, true, "0", duration.String(), "", t.monitor.MaxRetries); err != nil {
		fmt.Printf("Failed to insert monitor log: %v\n", err)
	}

	return nil
}

// runUdpMonitor 执行UDP监控
func (t *MonitorTask) runUdpMonitor() error {
	params, err := t.monitor.GetUdpTypeParams()
	if err != nil {
		// 插入错误监控日志
		if err := db.CreateMonitorLog(t.monitor.ID, false, "0", "0s", err.Error(), t.monitor.MaxRetries); err != nil {
			fmt.Printf("Failed to insert monitor log: %v\n", err)
		}

		return err
	}

	// 记录开始时间
	startTime := time.Now()

	// 连接UDP服务器
	addr := fmt.Sprintf("%s:%d", params.Host, params.Port)
	conn, err := net.DialTimeout("udp", addr, time.Duration(t.monitor.Timeout)*time.Second)
	if err != nil {
		// 插入错误监控日志
		if err := db.CreateMonitorLog(t.monitor.ID, false, "0", time.Since(startTime).String(), err.Error(), t.monitor.MaxRetries); err != nil {
			fmt.Printf("Failed to insert monitor log: %v\n", err)
		}

		return fmt.Errorf("UDP connection failed: %v", err)
	}
	defer conn.Close()

	// 发送测试数据
	testData := []byte("ping")
	_, err = conn.Write(testData)
	if err != nil {
		// 插入错误监控日志
		if err := db.CreateMonitorLog(t.monitor.ID, false, "0", time.Since(startTime).String(), err.Error(), t.monitor.MaxRetries); err != nil {
			fmt.Printf("Failed to insert monitor log: %v\n", err)
		}

		return fmt.Errorf("UDP write failed: %v", err)
	}

	// 设置读取超时
	conn.SetReadDeadline(time.Now().Add(time.Duration(t.monitor.Timeout) * time.Second))

	// 读取响应
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		// UDP可能没有响应，这是正常的
		fmt.Printf("UDP monitor for %s: %s:%d\n", t.monitor.Name, params.Host, params.Port)
		fmt.Printf("No response received (normal for UDP)\n")
	} else {
		fmt.Printf("UDP monitor for %s: %s:%d\n", t.monitor.Name, params.Host, params.Port)
		fmt.Printf("Response: %s\n", string(buf[:n]))
	}

	// 计算消耗时间
	duration := time.Since(startTime)
	fmt.Printf("Response time: %v\n", duration)
	fmt.Printf("UDP monitor %s: OK\n", t.monitor.Name)

	// 插入监控日志
	if err := db.CreateMonitorLog(t.monitor.ID, true, "0", duration.String(), "", t.monitor.MaxRetries); err != nil {
		fmt.Printf("Failed to insert monitor log: %v\n", err)
	}

	return nil
}

// InitMonitorask 初始化监控任务
func InitMonitorask() {
	fmt.Println("Starting to initialize monitor tasks...")
	manager := monitortask.GetManager()

	// 使用分页查询，支持大量数据
	pageSize := 100
	page := 1
	totalCount := 0
	addedCount := 0

	for {
		var monitors []model.Monitor
		offset := (page - 1) * pageSize

		// 只查询启用的监控，减少数据量
		if err := db.GetDb().Where("status = ?", true).Offset(offset).Limit(pageSize).Find(&monitors).Error; err != nil {
			fmt.Printf("Failed to get monitor list (page %d): %v\n", page, err)
			break
		}

		// 如果没有更多数据，退出循环
		if len(monitors) == 0 {
			break
		}

		fmt.Printf("Found %d monitors on page %d\n", len(monitors), page)
		totalCount += len(monitors)

		// 为每个监控创建任务
		for _, monitor := range monitors {
			fmt.Printf("Processing monitor: %s (ID: %d, Type: %s, Interval: %d)\n",
				monitor.Name, monitor.ID, monitor.Type, monitor.Interval)

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
			addedCount++
		}

		// 如果返回的数据少于页面大小，说明已经到了最后一页
		if len(monitors) < pageSize {
			break
		}

		page++
	}

	// 启动任务管理器
	manager.Start()
	fmt.Printf("Monitor tasks initialized: %d total, %d added\n", totalCount, addedCount)

	// 列出所有任务，确认所有任务都已添加
	tasks := manager.ListTasks()
	fmt.Printf("Total tasks added: %d\n", len(tasks))
	for _, task := range tasks {
		fmt.Printf("Task: %s (ID: %s, Cron: %s)\n", task.Name, task.ID, task.CronExpr)
	}
}
