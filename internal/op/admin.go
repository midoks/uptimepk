package op

import (
	// "fmt"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"uptimepk/internal/db"
	"uptimepk/internal/model"
	utils "uptimepk/internal/utils"
)

func InitAdmin(user string, pass string) error {
	_, err := db.GetAdminByID(1)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {

			salt := utils.RandString(16)
			admin := &model.Admin{
				Username:   user,
				Password:   model.TwoHashPwd(pass, salt),
				Salt:       salt,
				AllowLogin: true,
				Status:     true,
				SuperAdmin: true,
				FullName:   "超级管理员",
			}

			admin.CreateTime = time.Now().Unix()
			admin.UpdateTime = time.Now().Unix()
			if err := db.CreateAdmin(nil, admin); err != nil {
				return err
			}
		}
	}
	// fmt.Println("data:", data)
	return nil
}

func InitSetting() error {
	err := InitSettingAdminData()
	if err != nil {
		return err
	}

	err = InitSettingWebData()
	if err != nil {
		return err
	}
	return nil
}

func InitSettingAdminData() error {
	common_data := &model.SysSetting{
		Code: db.SettingAdminUI,
		Uid:  0,
	}

	common_data.SetAdminUIValue(model.SysSettingAdminUIValue{
		ProductName: "uptimepk",
		SystemName:  "监控面板",
	})
	common_data.UpdateTime = time.Now().Unix()
	_, err := db.GetSysSettingByCode(db.SettingAdminUI)
	if err == nil {
		if err := db.GetDb().Model(&model.SysSetting{}).Where("code = ?", db.SettingAdminUI).Updates(common_data).Error; err != nil {
			return err
		}
		return nil
	}

	common_data.CreateTime = time.Now().Unix()
	if err := db.GetDb().Create(common_data).Error; err != nil {
		return err
	}
	return nil
}

func InitSettingWebData() error {
	common_data := &model.SysSetting{
		Code: db.SettingWebUI,
		Uid:  0,
	}

	common_data.SetWebUIValue(model.SysSettingWebUIValue{
		Name:     "UPPK",
		Subtitle: "网站运行状态监控工具",
	})
	common_data.UpdateTime = time.Now().Unix()
	_, err := db.GetSysSettingByCode(db.SettingWebUI)
	if err == nil {
		if err := db.GetDb().Model(&model.SysSetting{}).Where("code = ?", db.SettingWebUI).Updates(common_data).Error; err != nil {
			return err
		}
		return nil
	}

	common_data.CreateTime = time.Now().Unix()
	if err := db.GetDb().Create(common_data).Error; err != nil {
		return err
	}
	return nil
}
