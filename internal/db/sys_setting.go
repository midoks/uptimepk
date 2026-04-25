package db

import (
	"sync"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"uptimepk/internal/model"
)

const (
	SettingAdminUI = "admin_ui"
	SettingWebUI   = "web_ui"
	SettingDbConf  = "db_conf"
	SettingLog     = "log_sys"
)

var (
	sysSettingCache    = make(map[string]*model.SysSetting)
	sysSettingMutex    sync.RWMutex
	sysSettingExpiry   = make(map[string]time.Time)
	sysSettingCacheDur = time.Minute * 5 // 缓存5分钟
)

func GetSysSettingByCode(code string) (*model.SysSetting, error) {
	sysSettingMutex.RLock()
	// 检查缓存是否存在且未过期
	if cached, found := sysSettingCache[code]; found {
		if time.Now().Before(sysSettingExpiry[code]) {
			sysSettingMutex.RUnlock()
			return cached, nil
		}
	}
	sysSettingMutex.RUnlock()

	// 缓存不存在或已过期，从数据库查询
	var u model.SysSetting
	if err := db.Where("code = ?", code).First(&u).Error; err != nil {
		return nil, errors.Wrapf(err, "failed get sys setting")
	}

	// 更新缓存
	sysSettingMutex.Lock()
	sysSettingCache[code] = &u
	sysSettingExpiry[code] = time.Now().Add(sysSettingCacheDur)
	sysSettingMutex.Unlock()

	return &u, nil
}

// ClearSysSettingCache 清除系统设置缓存
func ClearSysSettingCache() {
	sysSettingMutex.Lock()
	sysSettingCache = make(map[string]*model.SysSetting)
	sysSettingExpiry = make(map[string]time.Time)
	sysSettingMutex.Unlock()
}

func SysSettingDeleteByCode(tx *gorm.DB, code string) error {
	if tx == nil {
		tx = db
	}
	var d model.SysSetting
	return tx.Where("code = ?", code).Delete(&d).Error
}
