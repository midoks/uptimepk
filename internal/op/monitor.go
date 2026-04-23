package op

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

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
	// 获取HTTP监控参数
	params, err := t.monitor.GetHttpTypeParams()
	if err != nil {
		SysLog(err.Error())
		return err
	}

	// 记录重定向次数
	redirectCount := 0

	// 创建HTTP客户端，设置超时时间
	client := &http.Client{
		Timeout: time.Duration(t.monitor.Timeout) * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			redirectCount = len(via)
			if len(via) >= t.monitor.MaxRetries {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	// 初始化监控日志参数
	isValid := false
	size := 0
	var duration time.Duration
	errorMsg := ""
	var resp *http.Response

	// 发送HTTP请求
	startTime := time.Now()
	req, err := http.NewRequest("GET", params.Addr, nil)
	if err != nil {
		errorMsg = err.Error()
	} else {
		// 设置 User-Agent
		if params.UserAgent != "" {
			req.Header.Set("User-Agent", params.UserAgent)
		}
		resp, err = client.Do(req)
		if err != nil {
			errorMsg = err.Error()
		} else {
			defer resp.Body.Close()

			duration = time.Since(startTime)
			// 使用 io.LimitReader 限制读取大小，防止服务器发送过多数据
			maxSize := int64(10 * 1024 * 1024) // 10MB
			body, err := io.ReadAll(io.LimitReader(resp.Body, maxSize))
			if err != nil {
				// 忽略 Content-Length 不匹配的错误
				if !strings.Contains(err.Error(), "server replied with more than declared Content-Length") {
					errorMsg = err.Error()
				}
			}
			// 检查状态码
			isValid = resp.StatusCode >= 200 && resp.StatusCode < 300
			size = len(body)

			//检查内容
			if params.CheckContent != "" {
				bodyStr := string(body)
				if !strings.Contains(bodyStr, params.CheckContent) {
					isValid = false
					errorMsg = fmt.Sprintf("获取内容成功,但未匹配到字符串: %s", params.CheckContent)
				}
			}
		}
	}

	// 记录监控日志
	speedMs := 0.0
	if duration > 0 {
		speedMs = duration.Seconds() * 1000 // 转换为毫秒
	}
	if err := db.CreateMonitorLog(t.monitor.ID, isValid, size, speedMs, errorMsg, redirectCount); err != nil {
		return err
	}

	return nil
}

// runTcpMonitor 执行TCP监控
func (t *MonitorTask) runTcpMonitor() error {
	params, err := t.monitor.GetTcpTypeParams()
	if err != nil {
		SysLog(err.Error())
		return err
	}

	// 记录开始时间
	startTime := time.Now()

	// 连接TCP服务器
	addr := fmt.Sprintf("%s:%d", params.Host, params.Port)
	conn, err := net.DialTimeout("tcp", addr, time.Duration(t.monitor.Timeout)*time.Second)
	if err != nil {
		// 记录错误监控日志
		speedMs := time.Since(startTime).Seconds() * 1000 // 转换为毫秒
		if err := db.CreateMonitorLog(t.monitor.ID, false, 0, speedMs, err.Error(), t.monitor.MaxRetries); err != nil {
			fmt.Printf("failed to insert monitor log: %v\n", err)
		}
		return fmt.Errorf("TCP connection failed: %v", err)
	}
	defer conn.Close()

	// 计算消耗时间
	duration := time.Since(startTime)

	// 记录监控结果
	// fmt.Printf("TCP monitor for %s: %s:%d\n", t.monitor.Name, params.Host, params.Port)
	// fmt.Printf("Response time: %v\n", duration)
	// fmt.Printf("TCP monitor %s: OK\n", t.monitor.Name)

	// 记录监控日志
	speedMs := duration.Seconds() * 1000 // 转换为毫秒
	if err := db.CreateMonitorLog(t.monitor.ID, true, 0, speedMs, "", t.monitor.MaxRetries); err != nil {
		fmt.Printf("failed to insert monitor log: %v\n", err)
		return err
	}
	return nil
}

// runUdpMonitor 执行UDP监控
func (t *MonitorTask) runUdpMonitor() error {
	params, err := t.monitor.GetUdpTypeParams()
	if err != nil {
		SysLog(err.Error())
		return err
	}
	// 记录开始时间
	startTime := time.Now()

	// 连接UDP服务器
	addr := fmt.Sprintf("%s:%d", params.Host, params.Port)
	conn, err := net.DialTimeout("udp", addr, time.Duration(t.monitor.Timeout)*time.Second)
	if err != nil {
		speedMs := time.Since(startTime).Seconds() * 1000 // 转换为毫秒
		if err := db.CreateMonitorLog(t.monitor.ID, false, 0, speedMs, err.Error(), t.monitor.MaxRetries); err != nil {
			return err
		}
		return fmt.Errorf("udp connection failed: %v", err)
	}
	defer conn.Close()

	duration := time.Since(startTime)
	speedMs := duration.Seconds() * 1000 // 转换为毫秒
	if err := db.CreateMonitorLog(t.monitor.ID, true, 0, speedMs, "", t.monitor.MaxRetries); err != nil {
		return err
	}
	return nil
}

func MonitorAddTask(mm model.Monitor) error {
	mt_manager := monitortask.GetManager()

	// 根据监控间隔生成cron表达式
	// 例如：每60秒执行一次 -> "*/60 * * * * *"
	// 使用6字段cron表达式（秒、分、时、日、月、周）
	cronExpr := fmt.Sprintf("*/%d * * * * *", mm.Interval)
	task := &MonitorTask{monitor: &mm}

	// 添加任务到管理器
	if err := mt_manager.AddTask(task, cronExpr); err != nil {
		return fmt.Errorf("failed to add monitor task %s: %v\n", mm.Name, err)
	}
	return nil
}

func MonitorDeleteTask(mm model.Monitor) error {
	mt_manager := monitortask.GetManager()
	task := &MonitorTask{monitor: &mm}
	// 删除任务
	if err := mt_manager.RemoveTask(task.ID()); err != nil {
		return fmt.Errorf("failed to remove monitor task %s: %v\n", mm.Name, err)
	}
	return nil
}

func MonitorEnableTask(mm model.Monitor) error {
	mt_manager := monitortask.GetManager()
	task := &MonitorTask{monitor: &mm}
	// 删除任务
	if err := mt_manager.RemoveTask(task.ID()); err != nil {
		return fmt.Errorf("failed to enable monitor task %s: %v\n", mm.Name, err)
	}
	return nil
}

func MonitorDisableTask(mm model.Monitor) error {
	mt_manager := monitortask.GetManager()
	task := &MonitorTask{monitor: &mm}
	// 删除任务
	if err := mt_manager.DisableTask(task.ID()); err != nil {
		return fmt.Errorf("failed to disable monitor task %s: %v\n", mm.Name, err)
	}
	return nil
}

// InitMonitorask 初始化监控任务
func InitMonitorask() {
	fmt.Println("starting to initialize monitor tasks...")
	mt_manager := monitortask.GetManager()

	// 使用分页查询，支持大量数据
	pageSize := 100
	page := 1
	totalCount := 0
	addedCount := 0

	for {
		var monitors []model.Monitor
		offset := (page - 1) * pageSize

		// 只查询启用的监控，减少数据量
		if err := db.GetDb().Where("status = ?", true).Where("is_deleted = ?", 0).Offset(offset).Limit(pageSize).Find(&monitors).Error; err != nil {
			fmt.Printf("failed to get monitor list (page %d): %v\n", page, err)
			break
		}

		if len(monitors) == 0 {
			break
		}

		// fmt.Printf("Found %d monitors on page %d\n", len(monitors), page)
		totalCount += len(monitors)
		// 为每个监控创建任务
		for _, monitor := range monitors {
			if err := MonitorAddTask(monitor); err != nil {
				fmt.Printf("failed to add monitor task %s: %v\n", monitor.Name, err)
				continue
			}
			fmt.Printf("added monitor task %s with interval %d seconds\n", monitor.Name, monitor.Interval)
			addedCount++
		}

		// 如果返回的数据少于页面大小，说明已经到了最后一页
		if len(monitors) < pageSize {
			break
		}

		page++
	}

	// 启动任务管理器
	mt_manager.Start()
	fmt.Printf("Monitor tasks initialized: %d total, %d added\n", totalCount, addedCount)

	// 列出所有任务，确认所有任务都已添加
	tasks := mt_manager.ListTasks()
	fmt.Printf("Total tasks added: %d\n", len(tasks))
	for _, task := range tasks {
		fmt.Printf("Task: %s (ID: %s, Cron: %s)\n", task.Name, task.ID, task.CronExpr)
	}
}
