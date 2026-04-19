package db

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"

	log "github.com/sirupsen/logrus"

	"uptimepk/internal/model"
)

func GetMonitorList(page, size int) ([]model.Monitor, int64, error) {
	log.Info("GetMonitorList called with page:", page, " size:", size)
	cluster := db.Model(&model.Monitor{})
	var count int64
	if err := cluster.Count(&count).Error; err != nil {
		log.Error("Failed to count monitors:", err)
		return nil, 0, errors.Wrapf(err, "failed get server count")
	}
	log.Info("Monitor count:", count)

	var list []model.Monitor
	if err := db.Order(columnName("id")).Offset((page - 1) * size).Limit(size).Find(&list).Error; err != nil {
		log.Error("Failed to find monitors:", err)
		return nil, 0, errors.WithStack(err)
	}
	log.Info("Found monitors:", len(list))
	return list, count, nil
}

func GetMonitorByID(id int64) (*model.Monitor, error) {
	var u model.Monitor
	if err := db.First(&u, id).Error; err != nil {
		return nil, errors.Wrapf(err, "failed get admin")
	}
	return &u, nil
}

func MonitorDeleteByID(tx *gorm.DB, id int64) error {
	if tx == nil {
		tx = db
	}
	var d model.Monitor
	return tx.Where("id = ?", id).Delete(&d).Error
}
