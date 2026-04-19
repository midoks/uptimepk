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

func Database(c *gin.Context) {
	data := common.CommonVer(c)
	data["submenu"] = GetSysAdvancedSubMenu()
	c.HTML(http.StatusOK, "backend/system/database.tmpl", data)
}
