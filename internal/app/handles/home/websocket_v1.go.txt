package home

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"uptimepk/internal/db"
	"uptimepk/internal/model"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

const (
	MonitorBatchSize   = 20
	MonitorUpdateBatch = 10
)

type MonitorStatus struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Gid       int64     `json:"gid"`
	Type      string    `json:"type"`
	Status    bool      `json:"status"`
	IsValid   bool      `json:"is_valid"`
	Latency   string    `json:"latency"`
	Speed     float64   `json:"speed"`
	Size      int64     `json:"size"`
	ErrorMsg  string    `json:"error_msg"`
	HourLogs  []HourLog `json:"hour_logs"`
	UpRate    float64   `json:"up_rate"` // 今天的可用率
	UpdatedAt int64     `json:"updated_at"`
}

type HourLog struct {
	Hour     int64   `json:"hour"`
	Minute   int     `json:"minute"`
	IsValid  bool    `json:"is_valid"`
	ErrorMsg string  `json:"error_msg"`
	Speed    float64 `json:"speed"`
	Size     int64   `json:"size"`
}

type MonitorUpdate struct {
	ID        int64    `json:"id"`
	IsValid   bool     `json:"is_valid"`
	Latency   string   `json:"latency"`
	Speed     float64  `json:"speed"`
	Size      int64    `json:"size"`
	ErrorMsg  string   `json:"error_msg"`
	NewLog    *HourLog `json:"new_log,omitempty"`
	UpdatedAt int64    `json:"updated_at"`
}

type WSClient struct {
	conn          *websocket.Conn
	send          chan []byte
	isFirstUpdate bool
	lastUpdate    int64
}

type WSHub struct {
	clients    map[*WSClient]bool
	broadcast  chan []byte
	register   chan *WSClient
	unregister chan *WSClient
	mu         sync.RWMutex
}

var hub = &WSHub{
	clients:    make(map[*WSClient]bool),
	broadcast:  make(chan []byte),
	register:   make(chan *WSClient),
	unregister: make(chan *WSClient),
}

func init() {
	go hub.run()
	go broadcastLoop()
}

func (h *WSHub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
		}
	}
}

func (c *WSClient) readPump() {
	defer func() {
		hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("error: %v\n", err)
			}
			break
		}
	}
}

func (c *WSClient) writePump() {
	ticker := time.NewTicker(60 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte("\n"))
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func WSHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &WSClient{
		conn:          conn,
		send:          make(chan []byte, 256),
		isFirstUpdate: true,
		lastUpdate:    time.Now().Unix(),
	}
	hub.register <- client

	go client.writePump()
	go client.readPump()
}

func WSGroupsHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	// 发送分组数据
	var groups []model.MonitorGroup
	groupID := c.Query("id")

	if groupID != "" {
		// 如果指定了分组ID，只获取该分组
		groupIDInt, err := strconv.ParseInt(groupID, 10, 64)
		if err == nil {
			group, err := db.GetMonitorGroupByID(groupIDInt)
			if err == nil {
				groups = []model.MonitorGroup{*group}
			}
		}
	} else {
		// 否则获取所有分组
		groups, err = db.GetMonitorGroupAll()
		if err != nil {
			return
		}
	}

	// 获取每个分组的监控状态
	groupStatus := make([]map[string]interface{}, 0)
	for _, group := range groups {
		monitors, err := db.GetMonitorListByGid(group.ID)
		if err != nil {
			continue
		}

		monitorStatus := make([]MonitorStatus, 0)
		for _, monitor := range monitors {
			status, err := GetMonitorStatus(monitor.ID)
			if err != nil {
				continue
			}
			monitorStatus = append(monitorStatus, status)
		}

		groupData := map[string]interface{}{
			"id":       group.ID,
			"name":     group.Name,
			"monitors": monitorStatus,
		}
		groupStatus = append(groupStatus, groupData)
	}

	data := map[string]interface{}{
		"type": "group_monitors",
		"data": groupStatus,
	}

	message, err := json.Marshal(data)
	if err != nil {
		return
	}

	conn.WriteMessage(websocket.TextMessage, message)
}

func WSMonitorHandler(c *gin.Context) {
	monitorId := c.Query("id")
	if monitorId == "" {
		return
	}

	// 检查是否需要发送历史数据
	noHistory := c.Query("no_history") == "1"

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	// 发送监控点详情数据
	monitorIdInt, err := strconv.ParseInt(monitorId, 10, 64)
	if err != nil {
		return
	}
	monitor, err := db.GetMonitorByID(monitorIdInt)
	if err != nil {
		return
	}

	status, err := GetMonitorStatus(monitor.ID)
	if err != nil {
		return
	}

	// 获取分组名称
	group, err := db.GetMonitorGroupByID(monitor.Gid)
	groupName := ""
	if err == nil {
		groupName = group.Name
	}

	// 发送初始数据（今天的实时数据）
	initData := map[string]interface{}{
		"type":       "monitor_detail",
		"data":       status,
		"created_at": time.Unix(monitor.CreateTime, 0).Format("2006-01-02 15:04:05"),
		"updated_at": time.Unix(monitor.UpdateTime, 0).Format("2006-01-02 15:04:05"),
		"group_name": groupName,
	}

	initMessage, err := json.Marshal(initData)
	if err != nil {
		return
	}

	conn.WriteMessage(websocket.TextMessage, initMessage)

	// 如果不需要历史数据，直接进入实时更新循环
	if noHistory {
		// 发送完成消息（虽然没有历史数据，但也要告诉前端）
		doneData := map[string]interface{}{
			"type": "history_done",
		}
		doneMessage, _ := json.Marshal(doneData)
		conn.WriteMessage(websocket.TextMessage, doneMessage)

		// 进入实时更新循环
		client := &WSClient{
			conn:          conn,
			send:          make(chan []byte, 256),
			isFirstUpdate: true,
		}
		hub.register <- client
		go client.writePump()
		client.readPump()
		return
	}

	// 获取最近7天的日志（按天分组）
	weekLogs, err := db.GetMonitorLogsByDateRange(monitor.ID, time.Now().AddDate(0, 0, -7), time.Now())
	if err != nil {
		weekLogs = []model.MonitorLog{}
	}

	// 按天分组日志数据
	logsByDay := make(map[string][]map[string]interface{})
	for _, log := range weekLogs {
		dayKey := time.Unix(log.CreateTime, 0).Format("2006-01-02")
		logData := map[string]interface{}{
			"time":        time.Unix(log.CreateTime, 0).Format("15:04:05"),
			"is_valid":    log.IsValid,
			"error_msg":   log.ErrorMsg,
			"speed":       log.Speed,
			"size":        log.Size,
			"create_time": log.CreateTime, // 用于排序
		}
		logsByDay[dayKey] = append(logsByDay[dayKey], logData)
	}

	// 对每天内的日志按时间排序
	for day, logs := range logsByDay {
		sort.Slice(logs, func(i, j int) bool {
			return logs[i]["create_time"].(int64) < logs[j]["create_time"].(int64)
		})
		// 移除排序用的字段
		for i := range logs {
			delete(logs[i], "create_time")
		}
		logsByDay[day] = logs
	}

	// 按日期排序发送每天的数据
	type DayStat struct {
		Date      string                   `json:"date"`
		Total     int                      `json:"total"`
		UpCount   int                      `json:"up_count"`
		DownCount int                      `json:"down_count"`
		UpRate    float64                  `json:"up_rate"`
		Logs      []map[string]interface{} `json:"logs"`
	}

	var dayStats []DayStat
	for day, logs := range logsByDay {
		stat := DayStat{
			Date:  day,
			Total: len(logs),
			Logs:  logs,
		}
		for _, log := range logs {
			if log["is_valid"].(bool) {
				stat.UpCount++
			} else {
				stat.DownCount++
			}
		}
		if stat.Total > 0 {
			stat.UpRate = float64(stat.UpCount) / float64(stat.Total) * 100
		}
		dayStats = append(dayStats, stat)
	}

	// 发送每天的统计数据
	today := time.Now().Format("2006-01-02")
	for _, stat := range dayStats {
		// 跳过今天的数据
		if stat.Date == today {
			continue
		}
		data := map[string]interface{}{
			"type":       "history_day",
			"date":       stat.Date,
			"total":      stat.Total,
			"up_count":   stat.UpCount,
			"down_count": stat.DownCount,
			"up_rate":    stat.UpRate,
			"logs":       stat.Logs,
		}
		message, err := json.Marshal(data)
		if err != nil {
			continue
		}
		conn.WriteMessage(websocket.TextMessage, message)
		time.Sleep(100 * time.Millisecond)
	}

	// 发送完成消息
	doneData := map[string]interface{}{
		"type":       "history_done",
		"total_days": len(dayStats),
	}
	doneMessage, err := json.Marshal(doneData)
	if err != nil {
		return
	}
	conn.WriteMessage(websocket.TextMessage, doneMessage)

	// 进入实时更新循环
	client := &WSClient{
		conn:          conn,
		send:          make(chan []byte, 256),
		isFirstUpdate: true,
	}
	hub.register <- client
	go client.writePump()
	client.readPump()
}

func broadcastLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		BroadcastMonitorStatus()
	}
}

func GetMonitorStatus(monitorID int64) (MonitorStatus, error) {
	monitor, err := db.GetMonitorByID(monitorID)
	if err != nil {
		return MonitorStatus{}, err
	}

	status := MonitorStatus{
		ID:        monitor.ID,
		Name:      monitor.Name,
		Gid:       monitor.Gid,
		Type:      monitor.Type,
		Status:    monitor.Status != 0,
		IsValid:   false,
		Latency:   "",
		Speed:     0,
		Size:      0,
		ErrorMsg:  "",
		HourLogs:  GetMonitorHourLogs(monitorID),
		UpRate:    0,
		UpdatedAt: time.Now().Unix(),
	}

	latestLog, err := db.GetMonitorLatestLog(monitorID)
	if err == nil && latestLog != nil {
		status.IsValid = latestLog.IsValid
		status.Latency = fmt.Sprintf("%.2fms", latestLog.Speed)
		status.Speed = latestLog.Speed
		status.Size = latestLog.Size
		status.ErrorMsg = latestLog.ErrorMsg
	}

	// 计算今天的可用率
	today := time.Now()
	year, month, day := today.Date()
	todayInt := int64(year*10000 + int(month)*100 + day)
	logs, err := db.GetMonitorLogListByDate(monitorID, todayInt)
	if err == nil && len(logs) > 0 {
		upCount := 0
		for _, log := range logs {
			if log.IsValid {
				upCount++
			}
		}
		status.UpRate = float64(upCount) / float64(len(logs)) * 100
	}

	return status, nil
}

func GetMonitorHourLogs(monitorID int64) []HourLog {
	now := time.Now()
	year, month, day := now.Date()
	todayInt := int64(year*10000 + int(month)*100 + day)

	logs, err := db.GetMonitorLogListByDate(monitorID, todayInt)
	if err != nil {
		return []HourLog{}
	}

	hourLogs := make([]HourLog, 0, len(logs))
	for _, log := range logs {
		hourLogs = append(hourLogs, HourLog{
			Hour:     log.Hour,
			Minute:   log.Minute,
			IsValid:  log.IsValid,
			ErrorMsg: log.ErrorMsg,
			Speed:    log.Speed,
			Size:     log.Size,
		})
	}

	return hourLogs
}

func GetMonitorStatusList() ([]MonitorStatus, error) {
	monitors, _, err := db.GetMonitorListSimple(1, 100)
	if err != nil {
		return nil, err
	}

	statusList := make([]MonitorStatus, 0, len(monitors))
	for _, monitor := range monitors {
		status, err := GetMonitorStatus(monitor.ID)
		if err != nil {
			continue
		}
		statusList = append(statusList, status)
	}

	return statusList, nil
}

func GetMonitorUpdatesSince(since int64) ([]MonitorUpdate, error) {
	monitors, _, err := db.GetMonitorListSimple(1, 100)
	if err != nil {
		return nil, err
	}

	updates := make([]MonitorUpdate, 0, len(monitors))
	for _, m := range monitors {
		update := MonitorUpdate{
			ID:        m.ID,
			UpdatedAt: time.Now().Unix(),
		}

		latestLog, err := db.GetMonitorLatestLog(m.ID)
		if err == nil && latestLog != nil {
			update.IsValid = latestLog.IsValid
			update.Latency = fmt.Sprintf("%.2fms", latestLog.Speed)
			update.Speed = latestLog.Speed
			update.Size = latestLog.Size
			update.ErrorMsg = latestLog.ErrorMsg

			if latestLog.CreateTime > since {
				update.NewLog = &HourLog{
					Hour:     latestLog.Hour,
					Minute:   latestLog.Minute,
					IsValid:  latestLog.IsValid,
					ErrorMsg: latestLog.ErrorMsg,
					Speed:    latestLog.Speed,
					Size:     latestLog.Size,
				}
			}
		}

		updates = append(updates, update)
	}

	return updates, nil
}

func BroadcastMonitorStatus() {
	hub.mu.RLock()
	defer hub.mu.RUnlock()

	statusList, err := GetMonitorStatusList()
	if err != nil {
		return
	}

	total := len(statusList)
	if total == 0 {
		data := map[string]interface{}{
			"type":    "monitor_status",
			"data":    []MonitorStatus{},
			"groups":  nil,
			"updated": time.Now().Unix(),
			"total":   0,
			"chunk":   0,
			"chunks":  0,
		}

		if groups, err := db.GetMonitorGroupAll(); err == nil {
			data["groups"] = groups
		}

		message, err := json.Marshal(data)
		if err != nil {
			return
		}

		for client := range hub.clients {
			client.lastUpdate = time.Now().Unix()
			client.isFirstUpdate = false
			select {
			case client.send <- message:
			default:
				close(client.send)
				delete(hub.clients, client)
			}
		}
		return
	}

	chunks := (total + MonitorBatchSize - 1) / MonitorBatchSize

	for client := range hub.clients {
		client.lastUpdate = time.Now().Unix()
		client.isFirstUpdate = false

		if groups, err := db.GetMonitorGroupAll(); err == nil {
			for i := 0; i < total; i += MonitorBatchSize {
				end := i + MonitorBatchSize
				if end > total {
					end = total
				}
				chunkNum := i/MonitorBatchSize + 1

				data := map[string]interface{}{
					"type":    "monitor_status",
					"data":    statusList[i:end],
					"groups":  groups,
					"updated": time.Now().Unix(),
					"total":   total,
					"chunk":   chunkNum,
					"chunks":  chunks,
				}

				message, err := json.Marshal(data)
				if err != nil {
					continue
				}

				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(hub.clients, client)
				}

				if chunkNum < chunks {
					time.Sleep(10 * time.Millisecond)
				}
			}
		}
	}
}
