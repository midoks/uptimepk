package model

type AdminRecipientsMonitorRelated struct {
	ID          int64  `json:"id" gorm:"primaryKey"` // unique key
	RecipientID string `json:"recipient_id"`         // recipient_id
	MonitorGid  int64  `json:"monitor_gid"`          // monitor_gid
	Status      int    `json:"status"`               // status
	CreateTime  int64  `json:"create_time"`          // create_time
	UpdateTime  int64  `json:"update_time"`          // update_time
}
