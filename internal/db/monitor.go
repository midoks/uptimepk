package db

import (
	"time"

	"github.com/pkg/errors"

	"uptimepk/internal/app/entity"
	"uptimepk/internal/model"
)

func GetMonitorList(page, size int) ([]entity.MonitorEntityList, int64, error) {
	mm := db.Model(&model.Monitor{})
	var count int64
	if err := mm.Where("is_deleted=?", 0).Count(&count).Error; err != nil {
		return nil, 0, errors.Wrapf(err, "failed get monitor count")
	}

	var list []model.Monitor
	if err := db.Order(columnName("create_time")+" desc").Where("is_deleted=?", 0).Offset((page - 1) * size).Limit(size).Find(&list).Error; err != nil {
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

func GetMonitorListSimple(page, size int) ([]model.Monitor, int64, error) {
	cluster := db.Model(&model.Monitor{})
	var count int64
	if err := cluster.Count(&count).Error; err != nil {
		return nil, 0, errors.Wrapf(err, "failed get monitor count")
	}

	var list []model.Monitor
	if err := db.Order(columnName("id")).Offset((page - 1) * size).Limit(size).Find(&list).Error; err != nil {
		return nil, 0, errors.Wrap(err, "failed get monitor list")
	}
	return list, count, nil
}

func GetMonitorListByGid(gid int64) ([]model.Monitor, error) {
	var list []model.Monitor
	if err := db.Where("gid = ?", gid).Order(columnName("id")).Find(&list).Error; err != nil {
		return nil, errors.Wrap(err, "failed get monitor list by gid")
	}
	return list, nil
}

func GetMonitorByID(id int64) (*model.Monitor, error) {
	var u model.Monitor
	if err := db.First(&u, id).Error; err != nil {
		return nil, errors.Wrapf(err, "failed get monitor data")
	}
	return &u, nil
}

func MonitorTriggerStatus(id int64) error {

	var data model.Monitor
	if err := db.First(&data, id).Error; err != nil {
		return errors.Wrapf(err, "failed get monitor data")
	}

	var status bool
	if data.Status {
		status = false
	} else {
		status = true
	}

	if err := db.Model(&model.Monitor{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      status,
			"update_time": time.Now().Unix(),
		}).Error; err != nil {
		return err
	}
	return nil
}

func MonitorSoftDeleteByID(id int64) error {
	if err := db.Model(&model.Monitor{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_deleted":  1,
			"update_time": time.Now().Unix(),
		}).Error; err != nil {
		return err
	}
	return nil
}

func MonitorDeleteByID(id int64) error {
	var d model.Monitor
	return db.Where("id = ?", id).Delete(&d).Error
}
