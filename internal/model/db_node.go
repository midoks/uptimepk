package model

type DbNode struct {
	ID         int64  `json:"id" gorm:"primaryKey"` // unique key
	AdminID    int64  `json:"admin_id"`             // admin_id
	Name       string `json:"name"`                 // name
	Host       int64  `json:"host"`                 // host
	Port       int64  `json:"port"`                 // port
	Password   string `json:"password"`             // password
	Dbname     string `json:"db_name"`              // db_name
	Order      int64  `json:"order"`                // order
	Weigth     int64  `json:"weigth"`               // weigth
	IsOn       int    `json:"is_on"`                // is_on
	Status     bool   `json:"status"`               // status
	CreateTime int64  `json:"create_time"`          // create_time
	UpdateTime int64  `json:"update_time"`          // update_time
}
