package model

type Monitor struct {
	ID         int64  `json:"id" gorm:"primaryKey"` // unique key
	GID        string `json:"gid"`                  // gid
	Name       string `json:"name"`                 // name
	Type       string `json:"type"`                 // type
	Config     string `json:"config"`               // config
	Status     bool   `json:"status"`               // status
	CreateTime int64  `json:"create_time"`          // create_time
	UpdateTime int64  `json:"update_time"`          // update_time
}
