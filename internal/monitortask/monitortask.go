package monitortask

import (
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// Task 任务接口
type Task interface {
	Run() error
	ID() string
	Name() string
}

// TaskInfo 任务信息
type TaskInfo struct {
	ID        string
	Name      string
	CronExpr  string
	Task      Task
	EntryID   cron.EntryID
	Enabled   bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Manager 任务管理器
type Manager struct {
	cron    *cron.Cron
	tasks   map[string]*TaskInfo
	mutex   sync.RWMutex
	running bool
}

var taskManager *Manager

func init() {
	taskManager = &Manager{
		cron:  cron.New(cron.WithSeconds()),
		tasks: make(map[string]*TaskInfo),
	}
}

// GetManager 获取任务管理器实例
func GetManager() *Manager {
	return taskManager
}

// Start 启动任务管理器
func (m *Manager) Start() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.running {
		m.cron.Start()
		m.running = true
		fmt.Println("Task manager started")
	}
}

// Stop 停止任务管理器
func (m *Manager) Stop() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.running {
		m.cron.Stop()
		m.running = false
		fmt.Println("Task manager stopped")
	}
}

// 并发控制：限制同时运行的任务数量
var taskSemaphore chan struct{}

func init() {
	// 初始化任务信号量，限制并发任务数为 50
	taskSemaphore = make(chan struct{}, 50)
}

// AddTask 添加任务
func (m *Manager) AddTask(task Task, cronExpr string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	taskID := task.ID()
	if _, exists := m.tasks[taskID]; exists {
		return fmt.Errorf("task with id %s already exists", taskID)
	}

	// 验证 cron 表达式将在 AddFunc 中自动进行

	// 添加到 cron
	entryID, err := m.cron.AddFunc(cronExpr, func() {
		// 获取信号量，限制并发
		taskSemaphore <- struct{}{}
		defer func() {
			<-taskSemaphore
		}()

		// 执行任务，添加错误处理和超时控制
		done := make(chan error, 1)
		go func() {
			done <- task.Run()
		}()

		select {
		case err := <-done:
			if err != nil {
				fmt.Printf("Task %s failed: %v\n", task.Name(), err)
			}
		case <-time.After(30 * time.Second):
			fmt.Printf("Task %s timed out\n", task.Name())
		}
	})

	if err != nil {
		return fmt.Errorf("failed to add task to cron: %v", err)
	}

	taskInfo := &TaskInfo{
		ID:        taskID,
		Name:      task.Name(),
		CronExpr:  cronExpr,
		Task:      task,
		EntryID:   entryID,
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	m.tasks[taskID] = taskInfo

	// 如果管理器未运行，启动它
	if !m.running {
		// 直接启动，避免死锁
		m.cron.Start()
		m.running = true
	}
	// fmt.Printf("Task %s[%s] added with cron expression: %s\n", task.Name(), taskID, cronExpr)
	return nil
}

// RemoveTask 移除任务
func (m *Manager) RemoveTask(taskID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	taskInfo, exists := m.tasks[taskID]
	if !exists {
		return fmt.Errorf("task with id %s not found", taskID)
	}

	// 从 cron 中移除
	m.cron.Remove(taskInfo.EntryID)

	// 从任务列表中移除
	delete(m.tasks, taskID)
	// fmt.Printf("Task %s removed\n", taskInfo.Name)
	return nil
}

// UpdateTaskCron 更新任务的 cron 表达式
func (m *Manager) UpdateTaskCron(taskID string, cronExpr string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	taskInfo, exists := m.tasks[taskID]
	if !exists {
		return fmt.Errorf("task with id %s not found", taskID)
	}

	// 验证 cron 表达式将在 AddFunc 中自动进行

	// 移除旧的任务
	m.cron.Remove(taskInfo.EntryID)

	// 添加新的任务
	entryID, err := m.cron.AddFunc(cronExpr, func() {
		if err := taskInfo.Task.Run(); err != nil {
			fmt.Printf("Task %s failed: %v\n", taskInfo.Name, err)
		}
	})

	if err != nil {
		return fmt.Errorf("failed to update task cron: %v", err)
	}

	// 更新任务信息
	taskInfo.CronExpr = cronExpr
	taskInfo.EntryID = entryID
	taskInfo.UpdatedAt = time.Now()

	// fmt.Printf("Task %s[%d] cron updated to: %s\n", taskInfo.Name, entryID, cronExpr)
	return nil
}

// EnableTask 启用任务
func (m *Manager) EnableTask(taskID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	taskInfo, exists := m.tasks[taskID]
	if !exists {
		return fmt.Errorf("task with id %s not found", taskID)
	}

	if !taskInfo.Enabled {
		// 重新添加到 cron
		entryID, err := m.cron.AddFunc(taskInfo.CronExpr, func() {
			if err := taskInfo.Task.Run(); err != nil {
				fmt.Printf("Task %s failed: %v\n", taskInfo.Name, err)
			}
		})

		if err != nil {
			return fmt.Errorf("failed to enable task: %v", err)
		}

		taskInfo.EntryID = entryID
		taskInfo.Enabled = true
		taskInfo.UpdatedAt = time.Now()
		// fmt.Printf("Task %s enabled\n", taskInfo.Name)
	}

	return nil
}

// DisableTask 禁用任务
func (m *Manager) DisableTask(taskID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	taskInfo, exists := m.tasks[taskID]
	if !exists {
		return fmt.Errorf("task with id %s not found", taskID)
	}

	if taskInfo.Enabled {
		// 从 cron 中移除
		m.cron.Remove(taskInfo.EntryID)
		taskInfo.Enabled = false
		taskInfo.UpdatedAt = time.Now()
		fmt.Printf("Task %s disabled\n", taskInfo.Name)
	}

	return nil
}

// GetTask 获取任务信息
func (m *Manager) GetTask(taskID string) (*TaskInfo, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	taskInfo, exists := m.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task with id %s not found", taskID)
	}

	return taskInfo, nil
}

// ListTasks 列出所有任务
func (m *Manager) ListTasks() []*TaskInfo {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	tasks := make([]*TaskInfo, 0, len(m.tasks))
	for _, taskInfo := range m.tasks {
		tasks = append(tasks, taskInfo)
	}

	return tasks
}

// RemoveAllTasks 移除所有任务
func (m *Manager) RemoveAllTasks() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 停止 cron
	m.cron.Stop()

	// 清空任务列表
	for taskID, taskInfo := range m.tasks {
		fmt.Printf("Task %s removed\n", taskInfo.Name)
		delete(m.tasks, taskID)
	}

	// 重新创建 cron
	m.cron = cron.New(cron.WithSeconds())
	m.running = false

	fmt.Println("All tasks removed")
}
