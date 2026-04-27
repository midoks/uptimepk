package db

import (
	"strings"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"uptimepk/internal/app/form"
	"uptimepk/internal/model"
)

// applyLogFilters 应用日志查询过滤器
func applyLogFilters(query *gorm.DB, field form.LogList) *gorm.DB {
	// 条件查询: key like content
	if field.Key != "" {
		query = query.Where("content LIKE ?", "%"+field.Key+"%")
	}

	// 条件查询: times 时间范围
	if field.Times != "" {
		// 解析时间范围 "2026-04-27 00:00:00 - 2026-04-27 23:59:59"
		parts := strings.Split(field.Times, " - ")
		if len(parts) == 2 {
			startStr := strings.TrimSpace(parts[0])
			endStr := strings.TrimSpace(parts[1])

			// 验证时间字符串长度
			if len(startStr) >= 19 && len(endStr) >= 19 {
				// 解析时间格式 (带时区处理)
				const timeFormat = "2006-01-02 15:04:05"

				// 解析开始时间 (使用本地时区，与AddLog保持一致)
				start, err := time.ParseInLocation(timeFormat, startStr, time.Local)
				if err == nil {
					// 解析结束时间 (使用本地时区，与AddLog保持一致)
					end, err := time.ParseInLocation(timeFormat, endStr, time.Local)
					if err == nil {
						// 转换为 Unix 时间戳
						startUnix := start.Unix()
						endUnix := end.Unix()
						query = query.Where(columnName("create_time")+" BETWEEN ? AND ?", startUnix, endUnix)
					}
				}
			}
		}
	}

	return query
}

func GetLogListByArgs(field form.LogList) ([]model.Log, int64, error) {
	page := field.Page.Page
	size := field.Page.Limit

	// 应用过滤器
	baseQuery := applyLogFilters(db.Model(&model.Log{}), field)

	// 获取总数
	var count int64
	if err := baseQuery.Count(&count).Error; err != nil {
		return nil, 0, errors.Wrapf(err, "failed get server count")
	}

	// 获取分页数据
	var list []model.Log
	if err := baseQuery.Order(columnName("create_time") + " DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error; err != nil {
		return nil, 0, errors.WithStack(err)
	}

	return list, count, nil
}

func GetLogList(page, size int) ([]model.Log, int64, error) {
	mm := db.Model(&model.Log{})
	var count int64
	if err := mm.Count(&count).Error; err != nil {
		return nil, 0, errors.Wrapf(err, "failed get server count")
	}

	var list []model.Log
	if err := db.Order(columnName("id")).Offset((page - 1) * size).Limit(size).Find(&list).Error; err != nil {
		return nil, 0, errors.WithStack(err)
	}
	return list, count, nil
}

func GetLogByID(id int64) (*model.Log, error) {
	var u model.Log
	if err := db.First(&u, id).Error; err != nil {
		return nil, errors.Wrapf(err, "failed get log")
	}
	return &u, nil
}

func LogDeleteByID(tx *gorm.DB, id int64) error {
	if tx == nil {
		tx = db
	}
	var d model.Log
	return tx.Where("id = ?", id).Delete(&d).Error
}

func AddLog(tx *gorm.DB, uid int64, content string) error {
	if tx == nil {
		tx = db
	}
	var u model.Log
	u.Uid = uid
	u.Content = content
	u.CreateTime = time.Now().Unix()

	return errors.WithStack(tx.Create(&u).Error)
}

func LogDeleteAll(tx *gorm.DB) error {
	if tx == nil {
		tx = db
	}
	var d model.Log
	return errors.WithStack(tx.Where("1 = 1").Delete(&d).Error)
}

func LogDeleteBeforeDays(tx *gorm.DB, days int) error {
	if tx == nil {
		tx = db
	}
	if days <= 0 {
		return nil
	}
	cutoff := time.Now().Add(-time.Duration(days) * 24 * time.Hour).Unix()
	var d model.Log
	return errors.WithStack(tx.Where("create_time < ?", cutoff).Delete(&d).Error)
}
