package home

import (
	"encoding/json"
	"fmt"
	"net/http"
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

type MonitorStatus struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Gid       int64     `json:"gid"`
	Type      string    `json:"type"`
	Status    bool      `json:"status"`
	IsValid   bool      `json:"is_valid"`
	Latency   string    `json:"latency"`
	ErrorMsg  string    `json:"error_msg"`
	HourLogs  []HourLog `json:"hour_logs"`
	UpdatedAt int64     `json:"updated_at"`
}

type HourLog struct {
	Hour     int64   `json:"hour"`
	Minute   int     `json:"minute"`
	IsValid  bool    `json:"is_valid"`
	ErrorMsg string  `json:"error_msg"`
	Speed    float64 `json:"speed"`
}

type MonitorUpdate struct {
	ID        int64    `json:"id"`
	IsValid   bool     `json:"is_valid"`
	Latency   string   `json:"latency"`
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

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func broadcastLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		BroadcastMonitorUpdates()
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
		lastUpdate:    0,
	}

	hub.register <- client

	go client.writePump()
	go client.readPump()

	go func() {
		time.Sleep(100 * time.Millisecond)
		SendMonitorStatusToClient(client)
	}()
}

func (c *WSClient) readPump() {
	defer func() {
		hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		var req map[string]interface{}
		if err := json.Unmarshal(message, &req); err != nil {
			continue
		}

		if req["type"] == "ping" {
			c.send <- []byte(`{"type":"pong"}`)
		} else if req["type"] == "get_status" {
			BroadcastMonitorStatus()
		} else if req["type"] == "get_by_gid" {
			if gid, ok := req["gid"].(float64); ok {
				statusList := GetMonitorStatusListByGid(int64(gid))
				data := map[string]interface{}{
					"type":    "monitor_status_by_gid",
					"gid":     int64(gid),
					"data":    statusList,
					"updated": time.Now().Unix(),
				}
				if msg, err := json.Marshal(data); err == nil {
					c.send <- msg
				}
			}
		}
	}
}

func (c *WSClient) writePump() {
	ticker := time.NewTicker(30 * time.Second)
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

func GetMonitorStatusList() ([]MonitorStatus, error) {
	monitors, _, err := db.GetMonitorListSimple(1, 100)
	if err != nil {
		return nil, err
	}

	statusList := make([]MonitorStatus, 0, len(monitors))
	for _, m := range monitors {
		status := GetMonitorStatusFromMonitor(m)
		statusList = append(statusList, status)
	}

	return statusList, nil
}

func GetMonitorStatusListByGid(gid int64) []MonitorStatus {
	monitors, err := db.GetMonitorListByGid(gid)
	if err != nil {
		return []MonitorStatus{}
	}

	statusList := make([]MonitorStatus, 0, len(monitors))
	for _, m := range monitors {
		status := GetMonitorStatusFromMonitor(m)
		statusList = append(statusList, status)
	}

	return statusList
}

func GetMonitorStatusFromMonitor(m model.Monitor) MonitorStatus {
	status := MonitorStatus{
		ID:        m.ID,
		Name:      m.Name,
		Gid:       m.Gid,
		Type:      m.Type,
		Status:    m.Status,
		UpdatedAt: m.UpdateTime,
		HourLogs:  []HourLog{},
	}

	latestLog, err := db.GetMonitorLatestLog(m.ID)
	if err == nil && latestLog != nil {
		status.IsValid = latestLog.IsValid
		status.Latency = fmt.Sprintf("%.2fms", latestLog.Speed)
		status.ErrorMsg = latestLog.ErrorMsg
	}

	hourLogs := GetMonitorHourLogs(m.ID)
	status.HourLogs = hourLogs

	return status
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
		})
	}

	return hourLogs
}

func SendMonitorStatusToClient(client *WSClient) {
	statusList, err := GetMonitorStatusList()
	if err != nil {
		return
	}

	data := map[string]interface{}{
		"type":    "monitor_status",
		"data":    statusList,
		"groups":  nil,
		"updated": time.Now().Unix(),
	}

	if groups, err := db.GetMonitorGroupAll(); err == nil {
		data["groups"] = groups
	}

	message, err := json.Marshal(data)
	if err != nil {
		return
	}

	client.lastUpdate = time.Now().Unix()
	client.isFirstUpdate = false
	client.send <- message
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
			update.ErrorMsg = latestLog.ErrorMsg

			if latestLog.CreateTime > since {
				update.NewLog = &HourLog{
					Hour:     latestLog.Hour,
					Minute:   latestLog.Minute,
					IsValid:  latestLog.IsValid,
					ErrorMsg: latestLog.ErrorMsg,
					Speed:    latestLog.Speed,
				}
			}
		}

		updates = append(updates, update)
	}

	return updates, nil
}

func BroadcastMonitorUpdates() {
	hub.mu.RLock()
	defer hub.mu.RUnlock()

	for client := range hub.clients {
		var message []byte
		var err error

		if client.isFirstUpdate {
			statusList, err := GetMonitorStatusList()
			if err != nil {
				continue
			}

			data := map[string]interface{}{
				"type":    "monitor_status",
				"data":    statusList,
				"groups":  nil,
				"updated": time.Now().Unix(),
			}

			if groups, err := db.GetMonitorGroupAll(); err == nil {
				data["groups"] = groups
			}

			message, err = json.Marshal(data)
			client.isFirstUpdate = false
		} else {
			updates, err := GetMonitorUpdatesSince(client.lastUpdate)
			if err != nil {
				continue
			}

			data := map[string]interface{}{
				"type":    "monitor_updates",
				"data":    updates,
				"updated": time.Now().Unix(),
			}

			message, err = json.Marshal(data)
		}

		if err != nil {
			continue
		}

		client.lastUpdate = time.Now().Unix()
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(hub.clients, client)
		}
	}
}

func BroadcastMonitorStatus() {
	hub.mu.RLock()
	defer hub.mu.RUnlock()

	statusList, err := GetMonitorStatusList()
	if err != nil {
		return
	}

	data := map[string]interface{}{
		"type":    "monitor_status",
		"data":    statusList,
		"groups":  nil,
		"updated": time.Now().Unix(),
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
}
