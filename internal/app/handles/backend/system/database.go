package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"uptimepk/internal/app/common"
	"uptimepk/internal/app/form"
	// "uptimepk/internal/op"
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
			Name:   "修改模板",
			Link:   "system/database/update",
		},
		{
			Number: 3,
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
	c.HTML(http.StatusOK, "backend/system/database.tmpl", data)
}

func DatabaseUpdate(c *gin.Context) {
	data := common.CommonVer(c)
	data["submenu"] = GetSysAdvancedSubMenu()
	data["database_submenu"] = GetSysAdvancedDatabaseSubMenu()
	c.HTML(http.StatusOK, "backend/system/database_update.tmpl", data)
}

func DatabaseClean(c *gin.Context) {
	data := common.CommonVer(c)
	data["submenu"] = GetSysAdvancedSubMenu()
	data["database_submenu"] = GetSysAdvancedDatabaseSubMenu()
	c.HTML(http.StatusOK, "backend/system/database_cleans.tmpl", data)
}

func DatabaseCleanSetting(c *gin.Context) {
	data := common.CommonVer(c)
	data["submenu"] = GetSysAdvancedSubMenu()
	data["database_submenu"] = GetSysAdvancedDatabaseSubMenu()
	c.HTML(http.StatusOK, "backend/system/database_clean_setting.tmpl", data)
}
