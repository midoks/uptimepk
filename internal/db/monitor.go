package db

import (
	"fmt"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"uptimepk/internal/app/entity"
	"uptimepk/internal/model"
	"uptimepk/internal/monitortask"
)

// MonitorTask 监控任务
type MonitorTask struct {
	monitor *model.Monitor
}

// ID 获取任务ID
func (t *MonitorTask) ID() string {
	return "monitor_" + strconv.FormatInt(t.monitor.ID, 10)
}

// Name 获取任务名称
func (t *MonitorTask) Name() string {
	return t.monitor.Name
}

func GetMonitorList(page, size int) ([]entity.MonitorEntityList, int64, error) {
	cluster := db.Model(&model.Monitor{})
	var count int64
	if err := cluster.Count(&count).Error; err != nil {
		return nil, 0, errors.Wrapf(err, "failed get monitor count")
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
	mt_manager := monitortask.GetManager()
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

	task := &MonitorTask{monitor: &data}
	if status {
		if err := mt_manager.EnableTask(task.ID()); err != nil {
			return fmt.Errorf("failed to enable monitor task %s: %v\n", data.Name, err)
		}
	} else {
		if err := mt_manager.DisableTask(task.ID()); err != nil {
			return fmt.Errorf("failed to disable monitor task %s: %v\n", data.Name, err)
		}
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
