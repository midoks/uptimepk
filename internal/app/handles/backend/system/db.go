package server

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

func Db(c *gin.Context) {
	data := common.CommonVer(c)
	data["submenu"] = GetSysAdvancedSubMenu()
	c.HTML(http.StatusOK, "backend/system/db/index.tmpl", data)
}

func DbNodeAdd(c *gin.Context) {
	data := common.CommonVer(c)
	data["submenu"] = GetSysAdvancedSubMenu()
	c.HTML(http.StatusOK, "backend/system/db/add.tmpl", data)
}

func DbNodeDetails(c *gin.Context) {
	data := common.CommonVer(c)
	data["submenu"] = GetSysAdvancedSubMenu()

	id := c.Query("id")
	data["id"] = id

	idInt, _ := strconv.ParseInt(id, 10, 64)
	dbnode_data, err := db.GetDbNodeByID(idInt)
	if err == nil {
		data["Data"] = dbnode_data
	}

	c.HTML(http.StatusOK, "backend/system/db/details.tmpl", data)
}

func DbNodeClean(c *gin.Context) {
	data := common.CommonVer(c)
	data["submenu"] = GetSysAdvancedSubMenu()

	id := c.Query("id")
	data["id"] = id

	idInt, _ := strconv.ParseInt(id, 10, 64)
	dbnode_data, err := db.GetDbNodeByID(idInt)
	if err == nil {
		data["Data"] = dbnode_data
	}

	c.HTML(http.StatusOK, "backend/system/db/clean.tmpl", data)
}

func DbNodeUpdate(c *gin.Context) {
	data := common.CommonVer(c)
	data["submenu"] = GetSysAdvancedSubMenu()

	id := c.Query("id")
	data["id"] = id

	idInt, _ := strconv.ParseInt(id, 10, 64)
	dbnode_data, err := db.GetDbNodeByID(idInt)
	if err == nil {
		data["Data"] = dbnode_data
	}

	c.HTML(http.StatusOK, "backend/system/db/update.tmpl", data)
}

func DbNodeLogs(c *gin.Context) {
	data := common.CommonVer(c)
	data["submenu"] = GetSysAdvancedSubMenu()

	id := c.Query("id")
	data["id"] = id

	idInt, _ := strconv.ParseInt(id, 10, 64)
	dbnode_data, err := db.GetDbNodeByID(idInt)
	if err == nil {
		data["Data"] = dbnode_data
	}

	c.HTML(http.StatusOK, "backend/system/db/logs.tmpl", data)
}

func DbNodeList(c *gin.Context) {
	var field form.Page
	if err := c.ShouldBind(&field); err != nil {
		common.ErrorResp(c, err, -1)
		return
	}

	result, count, err := db.GetDbNodeList(field.Page, field.Limit)
	if err != nil {
		common.ErrorResp(c, err, -1)
		return
	}
	common.SuccessLayuiResp(c, count, "ok", result)
}

func PostDbNodeAdd(c *gin.Context) {
	var field form.DbNodeAdd
	var err error
	if err = c.ShouldBind(&field); err != nil {
		common.ErrorResp(c, err, -1)
		return
	}

	common_data := &model.DbNode{
		Name:       field.Name,
		Host:       field.Host, // 暂时设为0，需要根据实际情况转换
		Port:       int64(field.Port),
		Dbname:     field.Dbname,
		Username:   field.Username,
		Password:   field.Password,
		Order:      0, // 默认值
		Weigth:     0, // 默认值
		Status:     field.Status,
		UpdateTime: time.Now().Unix(),
	}

	if field.ID > 0 {
		if err := db.GetDb().Model(&model.DbNode{}).Where("id = ?", field.ID).Updates(common_data).Error; err != nil {
			common.ErrorResp(c, err, -1)
			return
		}
	} else {
		common_data.CreateTime = time.Now().Unix()
		if err := db.GetDb().Create(common_data).Error; err != nil {
			common.ErrorResp(c, err, -1)
			return
		}
	}
	common.SuccessResp(c)
}
