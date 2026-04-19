package model

type MonitorGroup struct {
	ID         int64  `json:"id" gorm:"primaryKey"` // unique key
	Name       string `json:"name"`                 // name
	RealTime   bool   `json:"real_time"`            // real_time
	Order      int64  `json:"order"`                // order
	Status     bool   `json:"status"`               // status
	CreateTime int64  `json:"create_time"`          // create_time
	UpdateTime int64  `json:"update_time"`          // update_time
}
