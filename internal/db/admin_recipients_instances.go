package db

import (
	"github.com/pkg/errors"

	"uptimepk/internal/model"
)

func GetAdminRecipientsInstancesList(page, size int) ([]model.AdminMediaInstance, int64, error) {
	adminM := db.Model(&model.AdminMediaInstance{})
	var count int64
	if err := adminM.Count(&count).Error; err != nil {
		return nil, 0, errors.Wrapf(err, "failed get recipients data list")
	}

	var list []model.AdminMediaInstance
	if err := db.Order(columnName("id")).Offset((page - 1) * size).Limit(size).Find(&list).Error; err != nil {
		return nil, 0, errors.WithStack(err)
	}
	return list, count, nil
}

func GetAdminRecipientsInstancesByID(id int64) (*model.AdminMediaInstance, error) {
	var u model.AdminMediaInstance
	if err := db.First(&u, id).Error; err != nil {
		return nil, errors.Wrapf(err, "failed get recipients data")
	}
	return &u, nil
}

func GetAdminRecipientsInstancesByTelegram() ([]model.AdminMediaInstance, error) {
	var list []model.AdminMediaInstance
	if err := db.Where("media_type = ?", "telegram").Find(&list).Error; err != nil {
		return nil, errors.Wrapf(err, "failed get telegram list data")
	}
	return list, nil
}

func AdminRecipientsInstancesDeleteByID(id int64) error {
	var d model.AdminMediaInstance
	return db.Where("id = ?", id).Delete(&d).Error
}
