package db

import (
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"uptimepk/internal/app/entity"
	"uptimepk/internal/model"
)

func GetMonitorList(page, size int) ([]entity.MonitorEntityList, int64, error) {
	cluster := db.Model(&model.Monitor{})
	var count int64
	if err := cluster.Count(&count).Error; err != nil {
		return nil, 0, errors.Wrapf(err, "failed get server count")
	}

	var list []model.Monitor
	if err := db.Order(columnName("id")).Offset((page - 1) * size).Limit(size).Find(&list).Error; err != nil {
		return nil, 0, errors.Wrap(err, "failed get monitor list")
	}

	if len(list) == 0 {
		return []entity.MonitorEntityList{}, count, nil
	}

	result := make([]entity.MonitorEntityList, len(list))
	for i, item := range list {
		loglist, _, _ := GetMonitorLogListByMonitorID(item.ID, 1, 10)
		result[i] = entity.MonitorEntityList{
			Monitor: item,
			LogList: loglist,
		}
	}

	return result, count, nil
}

func GetMonitorByID(id int64) (*model.Monitor, error) {
	var u model.Monitor
	if err := db.First(&u, id).Error; err != nil {
		return nil, errors.Wrapf(err, "failed get admin")
	}
	return &u, nil
}

func MonitorTriggerStatus(tx *gorm.DB, id int64) error {
	if tx == nil {
		tx = db
	}
	var data model.Monitor
	if err := tx.First(&data, id).Error; err != nil {
		return errors.Wrapf(err, "failed get monitor group")
	}

	var status bool
	if data.Status {
		status = false
	} else {
		status = true
	}

	if err := tx.Model(&model.Monitor{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      status,
			"update_time": time.Now().Unix(),
		}).Error; err != nil {
		return err
	}
	return nil
}

func MonitorDeleteByID(tx *gorm.DB, id int64) error {
	if tx == nil {
		tx = db
	}
	var d model.Monitor
	return tx.Where("id = ?", id).Delete(&d).Error
}
