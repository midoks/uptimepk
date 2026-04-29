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
	MonitorBatchSize = 20
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
	UpRate    float64   `json:"up_rate"` // 今天的可用率
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
		_, msg, err := c.conn.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("error: %v\n", err)
			}
			break
		}
		// fmt.Println("received message:", string(msg))

		// 解析 JSON 消息
		var message struct {
			Type      string `json:"type"`
			GroupID   int64  `json:"group_id"`
			MonitorID int64  `json:"monitor_id"`
			LastLogID int64  `json:"last_log_id"`
		}
		if err := json.Unmarshal(msg, &message); err != nil {
			fmt.Println("failed to parse message:", err)
			continue
		}

		if message.Type == "init_monitor_groups" {
			data := map[string]interface{}{
				"type": "init_monitor_groups",
			}
			if groups, err := db.GetMonitorGroupAll(); err == nil {
				data["groups"] = groups
			}
			responseMsg, _ := json.Marshal(data)
			c.conn.WriteMessage(websocket.TextMessage, responseMsg)
		} else if message.Type == "init_monitor_data" {
			data := map[string]interface{}{
				"type":       "init_monitor_data",
				"monitor_id": int64(message.MonitorID),
			}

			if message.MonitorID > 0 {
				status, err := GetMonitorStatusInit(message.MonitorID)
				// fmt.Println("status:", status)
				if err != nil {
					fmt.Println("failed to get monitor status init:", err)
					return
				}

				data["data"] = status
				initMessage, err := json.Marshal(data)
				if err != nil {
					fmt.Println("failed to get monitor status init list:", err)
					return
				}

				c.conn.WriteMessage(websocket.TextMessage, initMessage)
			} else {
				// 处理初始化监控数据
				statusList, err := GetMonitorStatusList()
				if err != nil {
					fmt.Println("failed to get monitor status list:", err)
					return
				}

				data["data"] = statusList

				initMessage, _ := json.Marshal(data)
				c.conn.WriteMessage(websocket.TextMessage, initMessage)
			}

		} else if message.Type == "append_monitor_data" {
			data := map[string]interface{}{
				"type":       "append_monitor_data",
				"monitor_id": int64(message.MonitorID),
			}

			todayInt := utils.TodayToDateInt()
			logs, err := db.GetMonitorLogListByDate(message.MonitorID, todayInt, message.LastLogID, 10)

			if err == nil {
				data["list"] = logs
			}

			responseMsg, _ := json.Marshal(data)
			c.conn.WriteMessage(websocket.TextMessage, responseMsg)
		} else if message.Type == "init_group_monitors" {
			// 发送分组数据
			var groups []model.MonitorGroup
			groupID := message.GroupID

			if groupID <= 0 {
				return
			}
			// 如果指定了分组ID，只获取该分组
			group, err := db.GetMonitorGroupByID(groupID)
			if err == nil {
				groups = []model.MonitorGroup{*group}
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
				groupStatus = append(groupStatus, groupData)
			}

			data := map[string]interface{}{
				"type": "init_group_monitors",
				"data": groupStatus,
			}

			message, err := json.Marshal(data)
			if err != nil {
				return
			}
			c.conn.WriteMessage(websocket.TextMessage, message)
		} else if message.Type == "init_history_day" {
			weekLogs, err := db.GetMonitorLogsByDateRangeByPos(message.MonitorID, time.Now().AddDate(0, 0, -7), time.Now(), 0, 10)
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
					"logs":       stat.Logs,
				}
				msg, err := json.Marshal(data)
				if err != nil {
					continue
				}
				c.conn.WriteMessage(websocket.TextMessage, msg)
				time.Sleep(100 * time.Millisecond)
			}

			// doneData := map[string]interface{}{
			// 	"type":       "init_history_day",
			// 	"total_days": len(dayStats),
			// }
			// doneMsg, err := json.Marshal(doneData)
			// if err != nil {
			// 	return
			// }
			// c.conn.WriteMessage(websocket.TextMessage, doneMsg)
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

	// 获取今天的日志（只查询一次）
	todayInt := utils.TodayToDateInt()
	logs, err := db.GetMonitorLogListByDate(monitorID, todayInt, 0, 10)

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

// GetMonitorHourLogsFromLogs 从已有的日志列表生成小时日志
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
	for _, monitor := range monitors {
		status, err := GetMonitorStatusInit(monitor.ID)
		if err != nil {
			continue
		}
		statusList = append(statusList, status)
	}

	return statusList, nil
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
