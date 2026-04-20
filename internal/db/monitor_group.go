package db

import (
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"uptimepk/internal/app/entity"
	"uptimepk/internal/model"
)

func GetMonitorGroupList(page, size int) ([]entity.MonitorGroupEntityList, int64, error) {
	var count int64
	if err := db.Model(&model.MonitorGroup{}).Count(&count).Error; err != nil {
		return nil, 0, errors.Wrapf(err, "failed get cluster group count")
	}

	var list []model.MonitorGroup
	if err := db.Order(columnName("order") + " ASC").Offset((page - 1) * size).Limit(size).Find(&list).Error; err != nil {
		return nil, 0, errors.WithStack(err)
	}

	if len(list) == 0 {
		return []entity.MonitorGroupEntityList{}, count, nil
	}

	// 收集所有分组的 ID
	groupIDs := make([]int64, len(list))
	for i, item := range list {
		groupIDs[i] = item.ID
	}

	// 使用子查询一次性获取所有分组的监控数量
	type groupCount struct {
		GID int64 `gorm:"column:gid"`
		Num int64 `gorm:"column:num"`
	}
	var counts []groupCount
	if err := db.Model(&model.Monitor{}).
		Select("gid, COUNT(*) as num").
		Where("gid IN ?", groupIDs).
		Group("gid").
		Find(&counts).Error; err != nil {
		return nil, 0, errors.Wrapf(err, "failed get monitor counts")
	}

	// 将计数转换为 map，方便查找
	countMap := make(map[int64]int64, len(counts))
	for _, c := range counts {
		countMap[c.GID] = c.Num
	}

	// 构建结果
	result := make([]entity.MonitorGroupEntityList, len(list))
	for i, item := range list {
		result[i] = entity.MonitorGroupEntityList{
			MonitorGroup: item,
			MonitorNum:   countMap[item.ID],
		}
	}
	return result, count, nil
}

func GetMonitorGroupAll() ([]model.MonitorGroup, error) {
	var list []model.MonitorGroup
	if err := db.Order(columnName("order")+" ASC").Where("status", 1).Find(&list).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return list, nil
}

func GetMonitorGroupAllByRelatedTg() ([]model.MonitorGroup, error) {
	var list []model.MonitorGroup
	if err := db.Order(columnName("order")+" ASC").Where("status", 1).Where("real_time", 1).Find(&list).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return list, nil
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
