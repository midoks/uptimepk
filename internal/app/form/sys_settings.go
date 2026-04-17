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
	ProductName string `form:"product_name" binding:"required"`
	SystemName  string `form:"system_name" binding:"required"`
}
