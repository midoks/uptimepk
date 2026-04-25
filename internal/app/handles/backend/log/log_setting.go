package log

import (
	"fmt"
	// "strconv"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"uptimepk/internal/app/common"
	"uptimepk/internal/app/form"
	"uptimepk/internal/db"
	"uptimepk/internal/model"
)

func Settings(c *gin.Context) {
	data := common.CommonVer(c)
	data["submenu"] = GetLogSubMenu()
	log_setting_data, err := db.GetSysSettingByCode(db.SettingLog)
	if err == nil {
		setting_data, _ := log_setting_data.GetLogValue()
		data["Data"] = setting_data
	}
	c.HTML(http.StatusOK, "backend/log/setting.tmpl", data)
}

func PostSettting(c *gin.Context) {
	var field form.LogSetting
	if err := c.ShouldBind(&field); err != nil {
		common.ErrorResp(c, err, 0)
		return
	}

	fmt.Println(field)
	fmt.Println("field.AllowedManualDelete:", field.AllowedManualDelete, "field.AllowedManual:", field.AllowedManual, "field.SaveDay:", field.SaveDay)

	common_data := &model.SysSetting{
		Code:       db.SettingLog,
		UpdateTime: time.Now().Unix(),
	}

	common_data.Uid = 0
	common_data.SetLogValue(model.SysSettingLogValue{
		AllowedManualDelete:   field.AllowedManualDelete,
		AllowedManual:         field.AllowedManual,
		SaveDay:               field.SaveDay,
		MaxCapacityLimit:      field.MaxCapacityLimit,
		MaxCapacityUnit:       field.MaxCapacityUnit,
		AllowedModClearConfig: field.AllowedModClearConfig,
	})

	_, err := db.GetSysSettingByCode(db.SettingLog)
	if err == nil {
		if err := db.GetDb().Model(&model.SysSetting{}).Where("code = ?", db.SettingLog).Updates(common_data).Error; err != nil {
			common.ErrorResp(c, err, -1)
			return
		}
		common.SuccessResp(c)
		return
	}

	common_data.CreateTime = time.Now().Unix()
	if err := db.GetDb().Create(common_data).Error; err != nil {
		common.ErrorResp(c, err, -1)
		return
	}
	common.SuccessResp(c)
}
