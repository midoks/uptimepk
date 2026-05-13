package install

import (
	// "fmt"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"uptimepk/internal/app/common"
	"uptimepk/internal/conf"
	"uptimepk/internal/db"
	"uptimepk/internal/op"
)

func HomePage(c *gin.Context) {
	step := c.Query("step")
	if step == "2" {
		Step2Page(c)
		return
	}
	data := common.CommonVer(c)
	c.HTML(http.StatusOK, "install/index.tmpl", data)
}

func Step2Page(c *gin.Context) {
	data := common.CommonVer(c)

	data["type"] = conf.Database.Type
	data["hostname"] = conf.Database.Hostname
	data["hostport"] = conf.Database.Hostport
	data["dbname"] = conf.Database.Name
	data["username"] = conf.Database.User
	data["password"] = conf.Database.Password
	data["table_prefix"] = conf.Database.TablePrefix
	data["dbpath"] = conf.Database.Path

	c.HTML(http.StatusOK, "install/step2.tmpl", data)
}

func MyDbtest(c *gin.Context) {
	install_data := make(map[string]string, 0)
	install_data["type"] = c.PostForm("type")
	install_data["hostname"] = c.PostForm("hostname")
	install_data["hostport"] = c.PostForm("hostport")
	install_data["dbname"] = c.PostForm("dbname")
	install_data["username"] = c.PostForm("username")
	install_data["password"] = c.PostForm("password")
	install_data["table_prefix"] = c.PostForm("table_prefix")
	install_data["dbpath"] = c.PostForm("dbpath")

	err := db.CheckDbConnnect(install_data)
	if err != nil {
		common.ErrorStrResp(c, "数据库连接失败: "+err.Error(), -1)
		return
	}

	c.JSON(200, common.Resp[interface{}]{
		Code: 200,
		Msg:  "数据库连接成功!",
	})
}

func PostInstallStep1(c *gin.Context) {
	install_data := make(map[string]string, 0)
	install_data["type"] = c.PostForm("type")
	install_data["hostname"] = c.PostForm("hostname")
	install_data["hostport"] = c.PostForm("hostport")
	install_data["dbname"] = c.PostForm("dbname")
	install_data["username"] = c.PostForm("username")
	install_data["password"] = c.PostForm("password")
	install_data["table_prefix"] = c.PostForm("table_prefix")
	install_data["dbpath"] = c.PostForm("dbpath")

	err := conf.InstallConf(install_data)
	if err != nil {
		common.ErrorStrResp(c, err.Error(), -1)
		return
	}

	init_account := c.PostForm("account")
	init_pass := c.PostForm("pass")

	if init_account == "" || init_pass == "" {
		common.ErrorStrResp(c, "初始化账户不能为为空!", -1)
		return
	}

	if conf.Security.InstallLock {
		db.InitDb()
		if err := op.InitAdmin(init_account, init_pass); err != nil {
			fmt.Println("InitAdmin error:", err)
		}
		if err := op.InitSetting(); err != nil {
			fmt.Println("InitSetting error:", err)
		}
	}

	c.JSON(200, common.Resp[interface{}]{
		Code: 200,
		Msg:  "安装成功!",
	})

}
