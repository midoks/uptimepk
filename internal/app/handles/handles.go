package handles

import (
	// "fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"uptimepk/internal/app/common"
	// "uptimepk/internal/op"
)

func AdminPage(c *gin.Context) {
	data := common.CommonVer(c)
	c.HTML(http.StatusOK, "index.tmpl", data)
}

func Home(c *gin.Context) {
	data := common.CommonVer(c)
	c.HTML(http.StatusOK, "home/index.tmpl", data)
}

func NotFound(c *gin.Context) {
	data := common.CommonVer(c)
	c.HTML(http.StatusOK, "404.tmpl", data)
}
