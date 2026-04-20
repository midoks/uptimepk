package db

import (
	"math"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"uptimepk/internal/model"
)

func GetMonitorLogList(page, size int) ([]model.MonitorLog, int64, error) {
	// 确保 page 至少为 1
	if page <= 0 {
		page = 1
	}
	// 确保 size 至少为 1
	if size <= 0 {
		size = 10
	}

	mmlog := GetDb().Model(&model.MonitorLog{})
	var count int64
	if err := mmlog.Count(&count).Error; err != nil {
		return nil, 0, errors.Wrapf(err, "failed get monitor log count")
	}

	var list []model.MonitorLog
	if err := GetDb().Order(columnName("id") + " desc").Offset((page - 1) * size).Limit(size).Find(&list).Error; err != nil {
		return nil, 0, errors.Wrapf(err, "failed get monitor log list")
	}
	return list, count, nil
}

// CreateMonitorLog 创建并插入监控日志
func CreateMonitorLog(monitorID int64, isValid bool, size int, speed float64, errorMsg string, maxRetries int) error {
	// 获取当前时间
	now := time.Now()
	year, month, day := now.Date()
	hour, minute, _ := now.Clock()
	timestamp := now.Unix()

	// 计算 yyyymmdd 格式的日期
	dayInt := year*10000 + int(month)*100 + day

	// speed 保留2位小数
	speed = math.Round(speed*100) / 100

	// 创建监控日志
	monitorLog := &model.MonitorLog{
		MonitorID:  strconv.FormatInt(monitorID, 10),
		Day:        int64(dayInt),
		Hour:       int64(hour),
		Minute:     minute,
		IsValid:    isValid,
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
