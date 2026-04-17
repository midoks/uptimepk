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

func GetSubMenu() []form.SubMenu {
	menu := []form.SubMenu{
		{
			Number: 1,
			Name:   "集群看板",
			Link:   "clusters/cluster/boards",
		},
		{
			Number: 2,
			Name:   "节点列表",
			Link:   "clusters/cluster/list",
		},
		{
			Number: 3,
			Name:   "创建节点",
			Link:   "clusters/cluster/create_node",
		},
		{
			Number: 4,
			Name:   "安装升级",
			Link:   "clusters/cluster/upgrade",
		},
		{
			Number: 5,
			Name:   "节点分组",
			Link:   "clusters/cluster/groups",
		},
		{
			Number: 6,
			Name:   "集群设置",
			Link:   "clusters/cluster/settings",
		},
		{
			Number: 7,
			Name:   "其它操作",
			Link:   "monitor/cluster/delete",
		},
	}
	return menu
}

func Home(c *gin.Context) {
	data := common.CommonVer(c)
	c.HTML(http.StatusOK, "backend/monitor/index.tmpl", data)
}

func Add(c *gin.Context) {
	data := common.CommonVer(c)
	c.HTML(http.StatusOK, "backend/monitor/add.tmpl", data)
}

func MonitorBoards(c *gin.Context) {
	data := common.CommonVer(c)
	data["submenu"] = GetSubMenu()
	data["cluster_id"] = c.Query("cluster_id")
	c.HTML(http.StatusOK, "backend/monitor/boards.tmpl", data)
}

func MonitorList(c *gin.Context) {
	data := common.CommonVer(c)
	data["submenu"] = GetSubMenu()
	data["cluster_id"] = c.Query("cluster_id")
	c.HTML(http.StatusOK, "backend/monitor/list.tmpl", data)
}

func MonitorDelete(c *gin.Context) {
	data := common.CommonVer(c)
	data["submenu"] = GetSubMenu()
	data["cluster_id"] = c.Query("cluster_id")
	c.HTML(http.StatusOK, "backend/monitor/delete.tmpl", data)
}

func Edit(c *gin.Context) {
	id := c.Query("id")
	idInt, _ := strconv.ParseInt(id, 10, 64)

	admin_data, _ := db.GetAdminByID(idInt)

	data := common.CommonVer(c)
	data["Data"] = admin_data
	c.HTML(http.StatusOK, "backend/cluster/edit.tmpl", data)
}

func List(c *gin.Context) {
	var field form.Page
	if err := c.ShouldBind(&field); err != nil {
		common.ErrorResp(c, err, -1)
		return
	}

	result, count, _ := db.GetMonitorList(field.Page, field.Limit)
	common.SuccessLayuiResp(c, count, "ok", result)
}

func PostCreate(c *gin.Context) {
	var field form.MonitorCreate
	if err := c.ShouldBind(&field); err != nil {
		common.ErrorResp(c, err, -1)
		return
	}

	monitor := &model.Monitor{
		Name:       field.Name,
		CreateTime: time.Now().Unix(),
		UpdateTime: time.Now().Unix(),
	}

	if err := db.GetDb().Create(monitor).Error; err != nil {
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

	err := db.MonitorDeleteByID(nil, field.ID)
	if err == nil {
		common.SuccessResp(c)
		return
	}
	common.ErrorResp(c, err, -1)
}
