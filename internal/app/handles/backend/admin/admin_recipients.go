package admin

import (
	// "encoding/json"
	// "fmt"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"uptimepk/internal/app/common"
	"uptimepk/internal/app/form"
	"uptimepk/internal/db"
	"uptimepk/internal/model"
	"uptimepk/internal/op"
)

func GetRecipientsSubMenu() []form.SubMenu {
	menu := []form.SubMenu{
		{
			Number: 1,
			Name:   "接收人",
			Link:   "admin/recipients",
		},
		{
			Number: 2,
			Name:   "接收人分组",
			Link:   "admin/recipients/groups",
		},
		{
			Number: 3,
			Name:   "媒介",
			Link:   "admin/recipients/instances",
		},
		// {
		// 	Number: 4,
		// 	Name:   "发送记录",
		// 	Link:   "admin/recipients/logs",
		// },
		// {
		// 	Number: 5,
		// 	Name:   "任务队列",
		// 	Link:   "admin/recipients/tasks",
		// },
	}
	return menu
}

// 通知媒介
func Recipients(c *gin.Context) {
	data := common.CommonVer(c)
	data["submenu"] = GetRecipientsSubMenu()
	c.HTML(http.StatusOK, "backend/admin/recipients/index.tmpl", data)
}

func RecipientsAdd(c *gin.Context) {
	data := common.CommonVer(c)

	monitorList, _, _ := db.GetMonitorGroupList(1, 100)
	data["MonitorList"] = monitorList

	admin_list, _, _ := db.GetAdminList(1, 100)
	data["AdminList"] = admin_list

	groupList, _, _ := db.GetAdminRecipientsGroupList(1, 100)
	data["GroupList"] = groupList

	recipients_list, _, _ := db.GetAdminRecipientsInstancesList(1, 100)
	data["RecipientsList"] = recipients_list

	recipients_monitor_related_list, _ := db.GetAdminRecipientsMonitorRelatedByRecipientID(0)
	data["RecipientsMonitorRelated"] = recipients_monitor_related_list

	c.HTML(http.StatusOK, "backend/admin/recipients/add.tmpl", data)
}

func PostRecipientsAdd(c *gin.Context) {
	var field form.AdminRecipients
	if err := c.ShouldBind(&field); err != nil {
		common.ErrorResp(c, err, -1)
		return
	}

	// if b, err := json.Marshal(field); err == nil {
	// 	fmt.Println(string(b))
	// } else {
	// 	fmt.Println("json marshal error:", err)
	// }

	if field.IntervalType == "second" {
		if !(field.Interval >= 5 && field.Interval <= 60) {
			common.ErrorResp(c, errors.New("选择秒时,范围之在[5-60]"), -1)
			return
		}
	} else if field.IntervalType == "minute" {
		if !(field.Interval >= 1 && field.Interval <= 60) {
			common.ErrorResp(c, errors.New("选择分钟时,范围之在[1-60]"), -1)
			return
		}
	} else if field.IntervalType == "hour" {
		if !(field.Interval >= 1 && field.Interval <= 23) {
			common.ErrorResp(c, errors.New("选择小时时,范围之在[1-23]"), -1)
			return
		}
	} else if field.IntervalType == "hour" {
		if !(field.Interval >= 1 && field.Interval <= 7) {
			common.ErrorResp(c, errors.New("选择小时时,范围之在[1-7]"), -1)
			return
		}
	}

	// 设置默认的 RelatedIDs RelatedIDs 不为空）
	common_data := &model.AdminRecipients{
		AdminID:      field.AdminID,
		MediaID:      field.MediaID,
		GroupID:      field.GroupID,
		RecipientID:  field.RecipientID,
		Interval:     field.Interval,
		IntervalType: field.IntervalType,
		Status:       field.Status,
		Mark:         field.Mark,
		UpdateTime:   time.Now().Unix(),
	}

	tx := db.GetDb().Begin()

	if field.ID > 0 {
		if err := tx.Model(&model.AdminRecipients{}).Where("id = ?", field.ID).Updates(common_data).Error; err != nil {
			tx.Rollback()
			common.ErrorResp(c, err, -1)
			return
		}

		if _, err := db.UpdateAdminRecipientsMonitorRelated(tx, field.ID, field.RelatedIDs); err != nil {
			tx.Rollback()
			common.ErrorResp(c, err, -1)
			return
		}
	} else {
		common_data.Status = true
		common_data.CreateTime = time.Now().Unix()
		if err := tx.Create(common_data).Error; err != nil {
			tx.Rollback()
			common.ErrorResp(c, err, -1)
			return
		}

		if _, err := db.UpdateAdminRecipientsMonitorRelated(tx, common_data.ID, field.RelatedIDs); err != nil {
			tx.Rollback()
			common.ErrorResp(c, err, -1)
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		common.ErrorResp(c, err, -1)
		return
	}

	// 重新加载接收人汇总任务
	go op.ReloadRecipientsSummaryTasks()

	common.SuccessResp(c)
}

func RecipientsList(c *gin.Context) {
	var field form.Page
	if err := c.ShouldBind(&field); err != nil {
		common.ErrorResp(c, err, -1)
		return
	}
	result, count, _ := db.GetAdminRecipientsList(field.Page, field.Limit)
	common.SuccessLayuiResp(c, count, "ok", result)
}

func RecipientsDelete(c *gin.Context) {
	var field form.ID
	if err := c.ShouldBind(&field); err != nil {
		common.ErrorResp(c, err, -1)
		return
	}

	err := db.AdminRecipientsDeleteByID(nil, field.ID)
	if err == nil {
		// 重新加载接收人汇总任务
		go op.ReloadRecipientsSummaryTasks()
		common.SuccessResp(c)
		return
	}
	common.ErrorResp(c, err, -1)
}
