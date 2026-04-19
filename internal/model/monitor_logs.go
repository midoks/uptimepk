package model

type MonitorLog struct {
	ID         int64  `json:"id" gorm:"primaryKey"` // unique key
	MonitorID  string `json:"monitor_id"`           // monitor_id
	Day        int64  `json:"day"`                  // day
	Hour       int64  `json:"hour"`                 // hour
	Minute     string `json:"minute"`               // minute
	IsConnect  bool   `json:"is_connect"`           // is_connect
	Speed      string `json:"speed"`                // speed
	ErrorMsg   string `json:"ErrorMsg"`             // params
	MaxRetries int    `json:"max_retries"`          // max_retries
	Mark       string `json:"mark"`                 // mark
	CreateTime int64  `json:"create_time"`          // create_time
	UpdateTime int64  `json:"update_time"`          // update_time
}
