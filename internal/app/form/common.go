package form

type SubMenu struct {
	Number int64  `form:"number"`
	Name   string `form:"name"`
	Link   string `form:"link"`
}

type SubSettingMenu struct {
	Number int64  `form:"number"`
	Name   string `form:"name"`
	Link   string `form:"link"`
	Type   string `form:"type"`
}

type Page struct {
	Page  int `form:"page"`
	Limit int `form:"limit"`
}

type ID struct {
	ID int64 `form:"id"`
}

type IDs struct {
	Ids string `json:"ids" binding:"required"`
}

type DatabaseCommon struct {
	TableName string `form:"table_name"`
}
