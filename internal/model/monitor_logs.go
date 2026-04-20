package model

type MonitorLog struct {
	ID         int64   `json:"id" gorm:"primaryKey"` // unique key
	MonitorID  string  `json:"monitor_id"`           // monitor_id
	Day        int64   `json:"day"`                  // day
	Hour       int64   `json:"hour"`                 // hour
	Minute     int     `json:"minute"`               // minute
	IsConnect  bool    `json:"is_connect"`           // is_connect
	Size       int64   `json:"size"`                 // size (byte)
	Speed      float64 `json:"speed"`                // speed (ms)
	ErrorMsg   string  `json:"error_msg"`            // error_msg
	MaxRetries int     `json:"max_retries"`          // max_retries
	CreateTime int64   `json:"create_time"`          // create_time
	UpdateTime int64   `json:"update_time"`          // update_time
}
