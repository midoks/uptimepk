package home

import (
	// "fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"uptimepk/internal/app/common"
	"uptimepk/internal/db"
	// "uptimepk/internal/op"
)

func Index(c *gin.Context) {
	data := common.FrontendCommonVer(c)

	groups, _ := db.GetMonitorGroupAll()
	data["groups"] = groups

	monitor_list, _, _ := db.GetMonitorListSimple(1, 10)
	data["monitor_list"] = monitor_list

	c.HTML(http.StatusOK, "home/index.tmpl", data)
}

func NotFound(c *gin.Context) {
	data := common.CommonVer(c)
	c.HTML(http.StatusOK, "404.tmpl", data)
}
