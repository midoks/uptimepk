package home

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"uptimepk/internal/db"
	"uptimepk/internal/model"
	"uptimepk/internal/utils"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

const (
	MonitorBatchSize = 100
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
	List      []ListLog `json:"list"`
	UpRate    float64   `json:"up_rate"`
	UpdatedAt int64     `json:"updated_at"`
}

type ListLog struct {
	ID       int64   `json:"id"`
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
	NewLog    *ListLog `json:"new_log,omitempty"`
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
	broadcast:  make(chan []byte, 1024),
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

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("websocket error: %v\n", err)
			}
			break
		}

		var message struct {
			Type      string `json:"type"`
			GroupID   int64  `json:"group_id"`
			MonitorID int64  `json:"monitor_id"`
			LastLogID int64  `json:"last_log_id"`
			Day       int64  `json:"day"`
		}

		if err := json.Unmarshal(msg, &message); err != nil {
			fmt.Printf("failed to parse message: %v, msg: %s\n", err, string(msg))
			continue
		}

		switch message.Type {
		case "init_monitor_groups":
			handleInitMonitorGroups(c)
		case "init_monitor_data":
			handleInitMonitorData(c, message.MonitorID)
		case "append_monitor_data":
			handleAppendMonitorData(c, message.MonitorID, message.LastLogID)
		case "init_group_monitors":
			handleInitGroupMonitors(c, message.GroupID)
		case "init_history_day":
			handleInitHistoryDay(c, message.MonitorID)
		case "append_history_data":
			handleAppendHistoryData(c, message.MonitorID, message.Day, message.LastLogID)
		}
	}
}

func handleInitMonitorGroups(c *WSClient) {
	data := map[string]interface{}{"type": "init_monitor_groups"}
	if groups, err := db.GetMonitorGroupAll(); err == nil {
		data["groups"] = groups
	}
	sendWSMessage(c, data)
}

func handleInitMonitorData(c *WSClient, monitorID int64) {
	data := map[string]interface{}{
		"type":       "init_monitor_data",
		"monitor_id": monitorID,
	}

	if monitorID > 0 {
		status, err := GetMonitorStatusInit(monitorID)
		if err != nil {
			fmt.Printf("failed to get monitor status init: %v\n", err)
			return
		}
		data["data"] = status
	} else {
		statusList, err := GetMonitorStatusList()
		if err != nil {
			fmt.Printf("failed to get monitor status list: %v\n", err)
			return
		}
		data["data"] = statusList
	}

	sendWSMessage(c, data)
}

func handleAppendMonitorData(c *WSClient, monitorID, lastLogID int64) {
	data := map[string]interface{}{
		"type":       "append_monitor_data",
		"monitor_id": monitorID,
	}

	todayInt := utils.TodayToDateInt()
	logs, err := db.GetMonitorLogListByDate(monitorID, todayInt, lastLogID, MonitorBatchSize)
	if err == nil {
		data["list"] = logs
	}

	sendWSMessage(c, data)
}

func handleInitGroupMonitors(c *WSClient, groupID int64) {
	if groupID <= 0 {
		return
	}

	group, err := db.GetMonitorGroupByID(groupID)
	if err != nil {
		fmt.Printf("failed to get monitor group: %v\n", err)
		return
	}

	monitors, err := db.GetMonitorListByGid(groupID)
	if err != nil {
		fmt.Printf("failed to get monitors by group: %v\n", err)
		return
	}

	monitorStatus := make([]MonitorStatus, 0, len(monitors))
	for _, monitor := range monitors {
		status, err := GetMonitorStatusInit(monitor.ID)
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

	data := map[string]interface{}{
		"type": "init_group_monitors",
		"data": []map[string]interface{}{groupData},
	}

	sendWSMessage(c, data)
}

func handleInitHistoryDay(c *WSClient, monitorID int64) {
	weekLogs, err := db.GetMonitorLogsByDateRangeByPos(monitorID, time.Now().AddDate(0, 0, -7), time.Now(), 0, MonitorBatchSize)
	if err != nil {
		weekLogs = []model.MonitorLog{}
	}

	logsByDay := make(map[string][]map[string]interface{})
	for _, log := range weekLogs {
		dayKey := time.Unix(log.CreateTime, 0).Format("2006-01-02")
		logData := map[string]interface{}{
			"time":        time.Unix(log.CreateTime, 0).Format("15:04:05"),
			"is_valid":    log.IsValid,
			"error_msg":   log.ErrorMsg,
			"speed":       log.Speed,
			"size":        log.Size,
			"create_time": log.CreateTime,
		}
		logsByDay[dayKey] = append(logsByDay[dayKey], logData)
	}

	for day, logs := range logsByDay {
		sort.Slice(logs, func(i, j int) bool {
			return logs[i]["create_time"].(int64) < logs[j]["create_time"].(int64)
		})
		for i := range logs {
			delete(logs[i], "create_time")
		}
		logsByDay[day] = logs
	}

	type DayStat struct {
		Date      string                   `json:"date"`
		Total     int                      `json:"total"`
		UpCount   int                      `json:"up_count"`
		DownCount int                      `json:"down_count"`
		UpRate    float64                  `json:"up_rate"`
		List      []map[string]interface{} `json:"list"`
	}

	var dayStats []DayStat
	for day, list := range logsByDay {
		stat := DayStat{
			Date:  day,
			Total: len(list),
			List:  list,
		}
		stat.UpCount, stat.DownCount = countValidLogs(list)
		if stat.Total > 0 {
			stat.UpRate = float64(stat.UpCount) / float64(stat.Total) * 100
		}
		dayStats = append(dayStats, stat)
	}

	today := time.Now().Format("2006-01-02")
	for _, stat := range dayStats {
		if stat.Date == today {
			continue
		}
		data := map[string]interface{}{
			"type":       "init_history_day",
			"date":       stat.Date,
			"total":      stat.Total,
			"up_count":   stat.UpCount,
			"down_count": stat.DownCount,
			"up_rate":    stat.UpRate,
			"list":       stat.List,
		}
		sendWSMessage(c, data)
		time.Sleep(100 * time.Millisecond)
	}

	doneData := map[string]interface{}{
		"type":       "history_done",
		"total_days": len(dayStats),
	}
	sendWSMessage(c, doneData)
}

func handleAppendHistoryData(c *WSClient, monitorID, day, lastLogID int64) {
	data := map[string]interface{}{
		"type":       "append_history_data",
		"monitor_id": monitorID,
		"day":        day,
	}

	logs, err := db.GetMonitorLogListByDate(monitorID, day, lastLogID, MonitorBatchSize)
	if err == nil {
		data["list"] = logs
	}

	sendWSMessage(c, data)
}

func sendWSMessage(c *WSClient, data map[string]interface{}) {
	msg, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("failed to marshal message: %v\n", err)
		return
	}
	if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
		fmt.Printf("failed to write message: %v\n", err)
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
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

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
		fmt.Printf("failed to upgrade connection: %v\n", err)
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

func broadcastLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		BroadcastMonitorStatus()
	}
}

func GetMonitorStatusInit(monitorID int64) (MonitorStatus, error) {
	monitor, err := db.GetMonitorByID(monitorID)
	if err != nil {
		return MonitorStatus{}, err
	}

	todayInt := utils.TodayToDateInt()
	logs, err := db.GetMonitorLogListByDate(monitorID, todayInt, 0, MonitorBatchSize)

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
		List:      GetMonitorHourLogsFromLogs(logs),
		UpRate:    calculateUpRate(logs),
		UpdatedAt: time.Now().Unix(),
	}

	// 优先从今天的日志中获取最新状态
	if len(logs) > 0 {
		latestLog := logs[len(logs)-1]
		status.IsValid = latestLog.IsValid
		status.Latency = fmt.Sprintf("%.2fms", latestLog.Speed)
		status.Speed = latestLog.Speed
		status.Size = latestLog.Size
		status.ErrorMsg = latestLog.ErrorMsg
	} else {
		// 如果今天没有日志，查询历史数据
		latestLog, err := db.GetMonitorLatestLog(monitorID)
		if err == nil && latestLog != nil {
			status.IsValid = latestLog.IsValid
			status.Latency = fmt.Sprintf("%.2fms", latestLog.Speed)
			status.Speed = latestLog.Speed
			status.Size = latestLog.Size
			status.ErrorMsg = latestLog.ErrorMsg
		}
	}

	return status, nil
}

func GetMonitorHourLogsFromLogs(logs []model.MonitorLog) []ListLog {
	listLogs := make([]ListLog, 0, len(logs))
	for _, log := range logs {
		listLogs = append(listLogs, ListLog{
			ID:       log.ID,
			Hour:     log.Hour,
			Minute:   log.Minute,
			IsValid:  log.IsValid,
			ErrorMsg: log.ErrorMsg,
			Speed:    log.Speed,
			Size:     log.Size,
		})
	}
	return listLogs
}

func GetMonitorStatusList() ([]MonitorStatus, error) {
	monitors, _, err := db.GetMonitorListSimple(1, 1000)
	if err != nil {
		return nil, err
	}

	statusList := make([]MonitorStatus, 0, len(monitors))
	todayInt := utils.TodayToDateInt()

	for i := range monitors {
		monitor := &monitors[i]
		status, err := getMonitorStatusWithCache(monitor, todayInt)
		if err != nil {
			continue
		}
		statusList = append(statusList, status)
	}

	return statusList, nil
}

// getMonitorStatusWithCache 带缓存的监控状态获取
func getMonitorStatusWithCache(monitor *model.Monitor, todayInt int64) (MonitorStatus, error) {
	logs, err := db.GetMonitorLogListByDate(monitor.ID, todayInt, 0, MonitorBatchSize)
	if err != nil {
		logs = []model.MonitorLog{}
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
		List:      GetMonitorHourLogsFromLogs(logs),
		UpRate:    calculateUpRate(logs),
		UpdatedAt: time.Now().Unix(),
	}

	// 优先从今天的日志中获取最新状态
	if len(logs) > 0 {
		latestLog := logs[len(logs)-1]
		status.IsValid = latestLog.IsValid
		status.Latency = fmt.Sprintf("%.2fms", latestLog.Speed)
		status.Speed = latestLog.Speed
		status.Size = latestLog.Size
		status.ErrorMsg = latestLog.ErrorMsg
	}

	return status, nil
}

func BroadcastMonitorStatus() {
	hub.mu.RLock()
	defer hub.mu.RUnlock()

	for client := range hub.clients {
		select {
		case client.send <- []byte("ping"):
		default:
			close(client.send)
			delete(hub.clients, client)
		}
	}
}

func calculateUpRate(logs []model.MonitorLog) float64 {
	if len(logs) == 0 {
		return 0
	}
	upCount := 0
	for _, log := range logs {
		if log.IsValid {
			upCount++
		}
	}
	return float64(upCount) / float64(len(logs)) * 100
}

func countValidLogs(logs []map[string]interface{}) (int, int) {
	up, down := 0, 0
	for _, log := range logs {
		if log["is_valid"].(bool) {
			up++
		} else {
			down++
		}
	}
	return up, down
}
