package admin

import (
	// "fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"uptimepk/internal/app/common"
	"uptimepk/internal/db"
)

func RecipientsRecipientsDetails(c *gin.Context) {
	id := c.Query("id")
	idint, _ := strconv.ParseInt(id, 10, 64)
	recipient_data, _ := db.GetAdminRecipientsByID(idint)

	data := common.CommonVer(c)
	data["submenu"] = GetRecipientsSubMenu()

	data["id"] = id
	data["Data"] = recipient_data

	// 获取关联的集群列表
	monitorRelatedList, _ := db.GetAdminRecipientsMonitorRelatedByRecipientID(idint)
	var monitorList []map[string]interface{}
	for _, related := range monitorRelatedList {
		mgid, _ := db.GetMonitorByID(related.MonitorGid)
		if mgid.ID > 0 {
			monitorList = append(monitorList, map[string]interface{}{
				"ID":   mgid.ID,
				"Name": mgid.Name,
			})
		}
	}
	data["MonitorList"] = monitorList

	c.HTML(http.StatusOK, "backend/admin/recipients/recipients_details.tmpl", data)
}

func RecipientsRecipientsUpdate(c *gin.Context) {
	id := c.Query("id")
	idint, _ := strconv.ParseInt(id, 10, 64)
	recipient_data, _ := db.GetAdminRecipientsByID(idint)

	data := common.CommonVer(c)
	data["id"] = id
	data["Data"] = recipient_data

	data["AdminID"] = recipient_data.AdminID
	data["MediaID"] = recipient_data.MediaID
	data["GroupID"] = recipient_data.GroupID

	monitorList, _, _ := db.GetMonitorGroupList(1, 100)
	data["MonitorList"] = monitorList

	admin_list, _, _ := db.GetAdminList(1, 100)
	data["AdminList"] = admin_list

	groupList, _, _ := db.GetAdminRecipientsGroupList(1, 100)
	data["GroupList"] = groupList

	recipients_list, _, _ := db.GetAdminRecipientsInstancesList(1, 100)
	data["RecipientsList"] = recipients_list

	recipients_monitor_related_list, _ := db.GetAdminRecipientsMonitorRelatedByRecipientID(idint)
	data["RecipientsMonitorRelated"] = recipients_monitor_related_list

	c.HTML(http.StatusOK, "backend/admin/recipients/recipients_update.tmpl", data)
}

func RecipientsRecipientsTest(c *gin.Context) {
	id := c.Query("id")
	idint, _ := strconv.ParseInt(id, 10, 64)
	recipient_data, _ := db.GetAdminRecipientsByID(idint)

	data := common.CommonVer(c)
	data["id"] = id
	data["Data"] = recipient_data
	c.HTML(http.StatusOK, "backend/admin/recipients/recipients_test.tmpl", data)
}
