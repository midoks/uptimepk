package monitor

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"uptimepk/internal/app/common"
	"uptimepk/internal/app/form"
	"uptimepk/internal/db"
	"uptimepk/internal/model"
)

func MonitorGroups(c *gin.Context) {
	data := common.CommonVer(c)
	c.HTML(http.StatusOK, "backend/monitor/groups/index.tmpl", data)
}

func MonitorGroupsAdd(c *gin.Context) {
	data := common.CommonVer(c)
	data["id"] = c.Query("id")
	if data["id"] != "" {
		qid, err := strconv.ParseInt(data["id"].(string), 10, 64)
		if err == nil {
			cg_data, err := db.GetMonitorGroupByID(qid)
			if err == nil {
				data["Data"] = cg_data
			}
		}
	}
	c.HTML(http.StatusOK, "backend/monitor/groups/add.tmpl", data)
}

func MonitorGroupsList(c *gin.Context) {
	result, count, _ := db.GetMonitorGroupList(1, 10)
	common.SuccessLayuiResp(c, count, "ok", result)
}

func MonitorGroupsTriggerStatus(c *gin.Context) {
	var field form.ID
	if err := c.ShouldBind(&field); err != nil {
		common.ErrorResp(c, err, -1)
		return
	}

	err := db.MonitorGroupTriggerStatus(nil, field.ID)
	if err == nil {
		common.SuccessResp(c)
		return
	}
	common.ErrorResp(c, err, -1)
}

func PostMonitorGroupsAdd(c *gin.Context) {
	var field form.MonitorGroupAdd
	if err := c.ShouldBind(&field); err != nil {
		common.ErrorResp(c, err, 0)
		return
	}

	common_data := &model.MonitorGroup{
		Name:   field.Name,
		Status: field.Status,
	}

	common_data.UpdateTime = time.Now().Unix()
	if field.ID != 0 {
		_, err := db.GetMonitorGroupByID(field.ID)
		if err == nil {
			if err := db.GetDb().Model(&model.MonitorGroup{}).Where("id = ?", field.ID).Updates(common_data).Error; err != nil {
				common.ErrorResp(c, err, -1)
				return
			}
			common.SuccessResp(c)
			return
		}
	}

	common_data.CreateTime = time.Now().Unix()
	if err := db.GetDb().Create(common_data).Error; err != nil {
		common.ErrorResp(c, err, -1)
		return
	}
	common.SuccessResp(c)
}

func MonitorGroupsDelete(c *gin.Context) {
	var field form.ID
	if err := c.ShouldBind(&field); err != nil {
		common.ErrorResp(c, err, -1)
		return
	}

	err := db.MonitorGroupDeleteByID(nil, field.ID)
	if err == nil {
		common.SuccessResp(c)
		return
	}
	common.ErrorResp(c, err, -1)
}
