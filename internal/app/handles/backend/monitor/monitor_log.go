package monitor

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"uptimepk/internal/app/common"
	"uptimepk/internal/app/form"
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

	// TODO: 实现 GetMonitorLogList 函数
	// result, count, _ := db.GetMonitorLogList(field.Page, field.Limit)
	var result []interface{}
	count := int64(0)
	common.SuccessLayuiResp(c, count, "ok", result)
}

func MonitorLogDelete(c *gin.Context) {
	var field form.ID
	if err := c.ShouldBind(&field); err != nil {
		common.ErrorResp(c, err, -1)
		return
	}

	// TODO: 实现 MonitorLogDeleteByID 函数
	// err := db.MonitorLogDeleteByID(nil, field.ID)
	// if err == nil {
	// 	common.SuccessResp(c)
	// 	return
	// }
	// common.ErrorResp(c, err, -1)
	common.SuccessResp(c)
}
