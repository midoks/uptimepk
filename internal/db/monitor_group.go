package db

import (
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"uptimepk/internal/app/entity"
	"uptimepk/internal/model"
)

func GetMonitorGroupList(page, size int) ([]entity.MonitorGroupEntityList, int64, error) {
	mm := db.Model(&model.MonitorGroup{})
	var count int64
	if err := mm.Count(&count).Error; err != nil {
		return nil, 0, errors.Wrapf(err, "failed get cluster group")
	}

	var list []model.MonitorGroup
	if err := db.Order(columnName("id")).Offset((page - 1) * size).Limit(size).Find(&list).Error; err != nil {
		return nil, 0, errors.WithStack(err)
	}

	if len(list) == 0 {
		return []entity.MonitorGroupEntityList{}, count, nil
	}

	result := make([]entity.MonitorGroupEntityList, 0, len(list))

	for _, item := range list {
		var num int64
		db.Model(&model.Monitor{}).Where("gid = ?", item.ID).Count(&num)
		entityItem := entity.MonitorGroupEntityList{
			MonitorGroup: item,
			MonitorNum:   num,
		}
		result = append(result, entityItem)
	}
	return result, count, nil
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
	var data model.MonitorGroup
	if err := tx.First(&data, id).Error; err != nil {
		return errors.Wrapf(err, "failed get monitor group")
	}

	var status bool
	if data.Status {
		status = false
	} else {
		status = true
	}

	if err := tx.Model(&model.MonitorGroup{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      status,
			"update_time": time.Now().Unix(),
		}).Error; err != nil {
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
