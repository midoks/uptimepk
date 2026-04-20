package db

import (
	"strconv"
	"time"

	"uptimepk/internal/model"
)

// CreateMonitorLog 创建并插入监控日志
func CreateMonitorLog(monitorID int64, isConnect bool, size int, speed float64, errorMsg string, maxRetries int) error {
	// 获取当前时间
	now := time.Now()
	year, month, day := now.Date()
	hour, minute, _ := now.Clock()
	timestamp := now.Unix()

	// 创建监控日志
	monitorLog := &model.MonitorLog{
		MonitorID:  strconv.FormatInt(monitorID, 10),
		Day:        int64(year*10000 + int(month)*100 + day),
		Hour:       int64(hour),
		Minute:     minute,
		IsConnect:  isConnect,
		Size:       int64(size),
		Speed:      speed,
		ErrorMsg:   errorMsg,
		MaxRetries: maxRetries,
		CreateTime: timestamp,
	}

	// 插入监控日志
	return GetDb().Create(monitorLog).Error
}
