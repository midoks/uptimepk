package server

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"uptimepk/internal/app/common"
	"uptimepk/internal/app/form"
	"uptimepk/internal/db"
	"uptimepk/internal/model"
)

func GetSysAdvancedSubMenu() []form.SubMenu {
	menu := []form.SubMenu{
		{
			Number: 1,
			Name:   "数据库",
			Link:   "system/database",
		},
		{
			Number: 2,
			Name:   "日志数据库",
			Link:   "system/db",
		},
	}
	return menu
}

func GetSysAdvancedDatabaseSubMenu() []form.SubMenu {
	menu := []form.SubMenu{
		{
			Number: 1,
			Name:   "配置模板",
			Link:   "system/database/index",
		},
		{
			Number: 2,
			Name:   "手动清理",
			Link:   "system/database/cleans",
		},
		{
			Number: 3,
			Name:   "自动清理设置",
			Link:   "system/database/clean_setting",
		},
	}
	return menu
}

func Database(c *gin.Context) {
	data := common.CommonVer(c)
	data["submenu"] = GetSysAdvancedSubMenu()
	data["database_submenu"] = GetSysAdvancedDatabaseSubMenu()
	c.HTML(http.StatusOK, "backend/system/database/index.tmpl", data)
}

func DatabaseUpdate(c *gin.Context) {
	data := common.CommonVer(c)
	data["submenu"] = GetSysAdvancedSubMenu()
	data["database_submenu"] = GetSysAdvancedDatabaseSubMenu()
	c.HTML(http.StatusOK, "backend/system/database/update.tmpl", data)
}

func DatabaseClean(c *gin.Context) {
	data := common.CommonVer(c)
	data["submenu"] = GetSysAdvancedSubMenu()
	data["database_submenu"] = GetSysAdvancedDatabaseSubMenu()
	c.HTML(http.StatusOK, "backend/system/database/cleans.tmpl", data)
}

func DatabaseCleanSetting(c *gin.Context) {
	data := common.CommonVer(c)
	data["submenu"] = GetSysAdvancedSubMenu()
	data["database_submenu"] = GetSysAdvancedDatabaseSubMenu()

	// 获取数据库配置设置
	setting_db_conf, err := db.GetSysSettingByCode(db.SettingDbConf)
	if err == nil {
		db_conf, _ := setting_db_conf.GetDbConfValue()
		data["monitor_log_days"] = db_conf.MonitorLogDays
	} else {
		data["monitor_log_days"] = 180
	}

	c.HTML(http.StatusOK, "backend/system/database/clean_setting.tmpl", data)
}

func PostDatabaseCleanSetting(c *gin.Context) {
	var field form.SettingDbConf
	var err error
	if err = c.ShouldBind(&field); err != nil {
		common.ErrorResp(c, err, -1)
		return
	}

	common_data := &model.SysSetting{
		Code: db.SettingDbConf,
		Uid:  0,
	}

	common_data.SetDbConfValue(model.SysSettingDbConfValue{
		MonitorLogDays: field.MonitorLogDays,
	})
	common_data.UpdateTime = time.Now().Unix()
	_, err = db.GetSysSettingByCode(db.SettingDbConf)
	if err == nil {
		if err := db.GetDb().Model(&model.SysSetting{}).Where("code = ?", db.SettingDbConf).Updates(common_data).Error; err != nil {
			common.ErrorResp(c, err, -1)
			return
		}
		db.ClearSysSettingCache()
		common.SuccessResp(c)
		return
	}

	common_data.CreateTime = time.Now().Unix()
	if err := db.GetDb().Create(common_data).Error; err != nil {
		common.ErrorResp(c, err, -1)
		return
	}
	common.SuccessResp(c)
}
