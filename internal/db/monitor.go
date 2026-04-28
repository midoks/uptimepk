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
	mm := db.Model(&model.Monitor{})
	var count int64
	if err := mm.Where("is_deleted=?", 0).Count(&count).Error; err != nil {
		return nil, 0, errors.Wrapf(err, "failed get monitor count")
	}

	var list []model.Monitor
	if err := db.Order(columnName("create_time")+" desc").Where("status=?", 1).Where("is_deleted=?", 0).Offset((page - 1) * size).Limit(size).Find(&list).Error; err != nil {
		return nil, 0, errors.Wrap(err, "failed get monitor list")
	}
	return list, count, nil
}

func GetMonitorListByGid(gid int64) ([]model.Monitor, error) {
	var list []model.Monitor
	if err := db.Where("gid = ?", gid).Where("is_deleted=?", 0).Order(columnName("id")).Find(&list).Error; err != nil {
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

// GetMonitorDeletedID 根据创建时间排序，获取一条删除的监控ID
func GetMonitorDeletedID() (int64, error) {
	var monitor model.Monitor
	if err := db.Order(columnName("create_time")).Where("is_deleted=?", 1).First(&monitor).Error; err != nil {
		return 0, errors.Wrap(err, "failed get deleted monitor")
	}
	return monitor.ID, nil
}

func GetMonitorByDeletedIDs() ([]int64, error) {
	var list []model.Monitor
	if err := db.Where("is_deleted=?", 1).Find(&list).Error; err != nil {
		return nil, errors.Wrap(err, "failed get deleted monitor list")
	}

	ids := make([]int64, len(list))
	for i, item := range list {
		ids[i] = item.ID
	}
	return ids, nil
}

func MonitorTriggerStatus(id int64) error {
	var data model.Monitor
	if err := db.First(&data, id).Error; err != nil {
		return errors.Wrapf(err, "failed get monitor data")
	}

	var status int
	if data.Status == 0 {
		status = 1
	} else {
		status = 0
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

func MonitorSoftDeleteByGid(gid int64) error {
	if err := db.Model(&model.Monitor{}).
		Where("gid = ?", gid).
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

// GetLatestMonitorID 根据创建时间排序，获取最新的一条监控数据的ID
func GetLatestMonitorID() (int64, error) {
	var monitor model.Monitor
	if err := db.Order(columnName("create_time") + " desc").First(&monitor).Error; err != nil {
		return 0, errors.Wrap(err, "failed get latest monitor")
	}
	return monitor.ID, nil
}

// GetOldestMonitorID 根据创建时间排序，获取最旧的一条监控数据的ID
func GetOldestMonitorID() (int64, error) {
	var monitor model.Monitor
	if err := db.Order(columnName("create_time") + " asc").First(&monitor).Error; err != nil {
		return 0, errors.Wrap(err, "failed get oldest monitor")
	}
	return monitor.ID, nil
}
