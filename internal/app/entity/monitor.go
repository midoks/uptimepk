package entity

import (
	"uptimepk/internal/model"
)

type MonitorGroupEntityList struct {
	model.MonitorGroup
	MonitorNum int64 `json:"monitor_num"` // 监控统计
}
