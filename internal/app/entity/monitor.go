package entity

import (
	"uptimepk/internal/model"
)

type MonitorEntityList struct {
	model.Monitor
	LogList []model.MonitorLog `json:"log_list"` // 监控日志
}

type MonitorGroupEntityList struct {
	model.MonitorGroup
	MonitorNum int64 `json:"monitor_num"` // 监控统计

}
