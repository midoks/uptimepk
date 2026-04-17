package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"uptimepk/internal/app/common"
	// "uptimepk/internal/app/form"
)

func Db(c *gin.Context) {
	data := common.CommonVer(c)
	data["submenu"] = GetSysAdvancedSubMenu()
	c.HTML(http.StatusOK, "backend/system/db.tmpl", data)
}
