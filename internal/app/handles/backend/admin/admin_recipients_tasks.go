package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"uptimepk/internal/app/common"
)

func RecipientsTasks(c *gin.Context) {
	data := common.CommonVer(c)
	data["submenu"] = GetRecipientsSubMenu()
	c.HTML(http.StatusOK, "backend/admin/recipients/tasks.tmpl", data)
}
