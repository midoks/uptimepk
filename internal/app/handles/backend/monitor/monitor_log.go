package monitor

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"uptimepk/internal/app/common"
	"uptimepk/internal/app/form"
	"uptimepk/internal/db"
)

func MonitorLog(c *gin.Context) {
	data := common.CommonVer(c)
	c.HTML(http.StatusOK, "backend/monitor/log/index.tmpl", data)
}

func MonitorLogList(c *gin.Context) {

	var field form.Page
	if err := c.ShouldBind(&field); err != nil {
		common.ErrorResp(c, err, -1)
		return
	}

	result, count, _ := db.GetMonitorLogList(field.Page, field.Limit)
	common.SuccessLayuiResp(c, count, "ok", result)
}

func MonitorLogDelete(c *gin.Context) {
	var field form.ID
	if err := c.ShouldBind(&field); err != nil {
		common.ErrorResp(c, err, -1)
		return
	}

	err := db.MonitorLogDeleteByID(nil, field.ID)
	if err == nil {
		common.SuccessResp(c)
		return
	}
	common.ErrorResp(c, err, -1)
}
