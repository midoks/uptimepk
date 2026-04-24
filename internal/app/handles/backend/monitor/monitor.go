package monitor

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"uptimepk/internal/app/common"
	"uptimepk/internal/app/form"
	"uptimepk/internal/db"
	"uptimepk/internal/model"
	"uptimepk/internal/monitortask"
	"uptimepk/internal/op"
)

func Home(c *gin.Context) {
	data := common.CommonVer(c)

	groups, _ := db.GetMonitorGroupAll()
	data["groups"] = groups

	c.HTML(http.StatusOK, "backend/monitor/index.tmpl", data)
}

func List(c *gin.Context) {
	var field form.Page
	if err := c.ShouldBind(&field); err != nil {
		common.ErrorResp(c, err, -1)
		return
	}

	result, count, err := db.GetMonitorList(field.Page, field.Limit)
	if err != nil {
		common.ErrorResp(c, err, -1)
		return
	}
	common.SuccessLayuiResp(c, count, "ok", result)
}

func Add(c *gin.Context) {
	data := common.CommonVer(c)
	data["id"] = c.Query("id")
	if data["id"] != "" {
		qid, err := strconv.ParseInt(data["id"].(string), 10, 64)
		if err == nil {
			monitor_data, err := db.GetMonitorByID(qid)
			if err == nil {
				data["Data"] = monitor_data
			}
		}
	}

	groups, _ := db.GetMonitorGroupAll()
	data["groups"] = groups
	c.HTML(http.StatusOK, "backend/monitor/add.tmpl", data)
}

func PostAdd(c *gin.Context) {
	var field form.MonitorAdd
	if err := c.ShouldBind(&field); err != nil {
		common.ErrorResp(c, err, -1)
		return
	}

	is_create := true

	common_data := &model.Monitor{
		Name:       field.Name,
		Type:       field.Type,
		Status:     field.Status,
		Interval:   field.Interval,
		MaxRetries: field.MaxRetries,
		Timeout:    field.Timeout,
		Gid:        field.Gid,
		CreateTime: time.Now().Unix(),
		UpdateTime: time.Now().Unix(),
	}

	if field.Type == "http" {
		common_data.SetHttpTypeParams(model.MonitorHttpTypeParams{
			Addr:         field.Addr,
			CheckContent: field.CheckContent,
			UserAgent:    field.UserAgent,
		})
	}

	if field.Type == "tcp" {
		common_data.SetTcpTypeParams(model.MonitorTcpTypeParams{
			Host: field.TcpHost,
			Port: field.TcpPort,
		})
	}

	if field.Type == "udp" {
		common_data.SetUdpTypeParams(model.MonitorUdpTypeParams{
			Host: field.UdpHost,
			Port: field.UdpPort,
		})
	}

	if field.ID != 0 {
		_, err := db.GetMonitorByID(field.ID)
		if err == nil {
			if err := db.GetDb().Model(&model.Monitor{}).Where("id = ?", field.ID).Updates(common_data).Error; err != nil {
				common.ErrorResp(c, err, -1)
				return
			}
			is_create = false
			common.SuccessResp(c)
			return
		}
	}

	if err := db.GetDb().Create(common_data).Error; err != nil {
		common.ErrorResp(c, err, -1)
		return
	}

	// 添加/更新任务
	if is_create && field.Status { // 新创建的并开启，加入计划任务
		op.MonitorAddTask(*common_data)
	} else if !is_create && field.Status { // 更新并开启，加入计划任务
		op.MonitorDeleteTask(*common_data)
		op.MonitorAddTask(*common_data)
	}
	common.SuccessResp(c)
}

func MonitorTriggerStatus(c *gin.Context) {
	var field form.ID
	if err := c.ShouldBind(&field); err != nil {
		common.ErrorResp(c, err, -1)
		return
	}

	err := db.MonitorTriggerStatus(field.ID)
	if err != nil {
		common.ErrorResp(c, err, -1)
		return
	}

	var data model.Monitor
	if err := db.GetDb().First(&data, field.ID).Error; err != nil {
		common.ErrorResp(c, err, -1)
		return
	}

	// 计划任务
	if data.Status {
		if err := op.MonitorAddTask(data); err != nil {
			common.ErrorResp(c, err, -1)
			return
		}

	} else {
		if err := op.MonitorDeleteTask(data); err != nil {
			common.ErrorResp(c, err, -1)
			return
		}
	}
	common.SuccessResp(c)
}

func MonitorReloadTask(c *gin.Context) {
	// 重新加载所有监控任务
	if err := op.MonitorReloadTask(); err != nil {
		common.ErrorResp(c, err, -1)
		return
	}
	common.SuccessResp(c)
}
func SoftDelete(c *gin.Context) {
	var field form.ID
	if err := c.ShouldBind(&field); err != nil {
		common.ErrorResp(c, err, -1)
		return
	}

	err := db.MonitorSoftDeleteByID(field.ID)
	if err != nil {
		common.ErrorResp(c, err, -1)
		return
	}

	var data model.Monitor
	if err := db.GetDb().First(&data, field.ID).Error; err != nil {
		common.ErrorResp(c, err, -1)
		return
	}

	// 删除任务
	if err := op.MonitorDeleteTask(data); err != nil {
		common.ErrorResp(c, err, -1)
		return
	}
	common.SuccessResp(c)
}

func Delete(c *gin.Context) {
	var field form.ID
	if err := c.ShouldBind(&field); err != nil {
		common.ErrorResp(c, err, -1)
		return
	}

	err := db.MonitorDeleteByID(field.ID)
	if err != nil {
		common.ErrorResp(c, err, -1)
		return
	}

	// 删除任务
	mt_manager := monitortask.GetManager()
	taskID := fmt.Sprintf("monitor_%d", field.ID)
	mt_manager.RemoveTask(taskID)

	common.SuccessResp(c)
}
