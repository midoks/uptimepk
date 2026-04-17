package db

import (
	"time"

	"gorm.io/gorm"

	"uptimepk/internal/model"

	"github.com/pkg/errors"
)

func GetMonitorGroupList(page, size int) ([]model.MonitorGroup, int64, error) {
	cluster := db.Model(&model.MonitorGroup{})
	var count int64
	if err := cluster.Count(&count).Error; err != nil {
		return nil, 0, errors.Wrapf(err, "failed get cluster group")
	}

	var list []model.MonitorGroup
	if err := db.Order(columnName("id")).Offset((page - 1) * size).Limit(size).Find(&list).Error; err != nil {
		return nil, 0, errors.WithStack(err)
	}
	return list, count, nil
}

func AddMonitorGroup(tx *gorm.DB, name string, clusterId int64) error {
	if tx == nil {
		tx = db
	}
	data := &model.MonitorGroup{
		Name: name,
	}

	data.CreateTime = time.Now().Unix()
	data.UpdateTime = time.Now().Unix()
	if err := errors.WithStack(tx.Create(data).Error); err != nil {
		return err
	}
	return nil
}

func MonitorGroupTriggerStatus(tx *gorm.DB, id int64) error {
	if tx == nil {
		tx = db
	}
	var data model.Admin
	if err := tx.First(&data, id).Error; err != nil {
		return errors.Wrapf(err, "failed get monitor group")
	}

	var status bool
	if data.Status {
		status = false
	} else {
		status = true
	}

	data.UpdateTime = time.Now().Unix()
	data.Status = status

	if err := tx.Model(&model.Admin{}).
		Where("id = ?", id).
		Updates(&data).Error; err != nil {
		return err
	}
	return nil
}

func UpdateMonitorGroup(tx *gorm.DB, name string, id int64) error {
	if tx == nil {
		tx = db
	}
	data := &model.MonitorGroup{
		Name: name,
	}

	data.UpdateTime = time.Now().Unix()
	if err := tx.Model(&model.MonitorGroup{}).
		Where("id = ?", id).
		Updates(&data).Error; err != nil {
		return err
	}
	return nil
}

func GetMonitorGroupByID(id int64) (*model.MonitorGroup, error) {
	var data model.MonitorGroup
	if err := db.First(&data, id).Error; err != nil {
		return nil, errors.Wrapf(err, "failed get cluster group")
	}
	return &data, nil
}

func MonitorGroupDeleteByID(tx *gorm.DB, id int64) error {
	if tx == nil {
		tx = db
	}
	var d model.MonitorGroup
	return tx.Where("id = ?", id).Delete(&d).Error
}
