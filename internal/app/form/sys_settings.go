package form

type SettingProfile struct {
	Name string `form:"name"`
}

type SettingLogin struct {
	Name      string `form:"name"`
	Password  string `form:"password"`
	Password2 string `form:"password2"`
}

type SettingAdminUI struct {
	DomainName  string `form:"domain_name" binding:"required"`
	ProductName string `form:"product_name" binding:"required"`
	SystemName  string `form:"system_name" binding:"required"`
}

type SettingWebUI struct {
	Name     string `form:"name"`
	Subtitle string `form:"subtitle"`
}

type SettingDbConf struct {
	MonitorLogDays int64 `form:"monitor_log_days"`
}
