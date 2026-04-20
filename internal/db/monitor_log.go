package db

import (
	"strconv"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"uptimepk/internal/model"
)

func GetMonitorLogList(page, size int) ([]model.MonitorLog, int64, error) {
	cluster := db.Model(&model.MonitorLog{})
	var count int64
	if err := cluster.Count(&count).Error; err != nil {
		return nil, 0, errors.Wrapf(err, "failed get server count")
	}

	var list []model.MonitorLog
	if err := db.Order(columnName("id")).Offset((page - 1) * size).Limit(size).Find(&list).Error; err != nil {
		return nil, 0, errors.WithStack(err)
	}
	return list, count, nil
}

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

func MonitorLogDeleteByID(tx *gorm.DB, id int64) error {
	if tx == nil {
		tx = GetDb()
	}
	var d model.MonitorLog
	return tx.Where("id = ?", id).Delete(&d).Error
}
