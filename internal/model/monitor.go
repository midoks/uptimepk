package model

import (
	"encoding/json"
	"fmt"
)

type Monitor struct {
	ID           int64  `json:"id" gorm:"primaryKey"`                  // unique key
	Gid          int64  `json:"gid"`                                   // gid
	Name         string `json:"name"`                                  // name
	Type         string `json:"type"`                                  // type
	Params       string `json:"params"`                                // params
	Status       int    `json:"status"`                                // status
	Interval     int    `json:"interval" gorm:"default:60"`            // interval
	IntervalType string `json:"interval_type" gorm:"default:'second'"` // interval_type
	MaxRetries   int    `json:"max_retries"`                           // max_retries
	Timeout      int    `json:"timeout"`                               // timeout
	Mark         string `json:"mark"`                                  // mark
	IsDeleted    int    `json:"is_deleted" gorm:"default:0"`           // is_deleted
	CreateTime   int64  `json:"create_time"`                           // create_time
	UpdateTime   int64  `json:"update_time"`                           // update_time
}

type MonitorHttpTypeParams struct {
	Addr         string `json:"addr"`
	CheckContent string `json:"check_content"`
	UserAgent    string `json:"user_agent"`
}

type MonitorTcpTypeParams struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type MonitorUdpTypeParams struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

func (a *Monitor) SetHttpTypeParams(p MonitorHttpTypeParams) error {
	b, err := json.Marshal(p)
	if err != nil {
		return err
	}
	a.Params = string(b)
	return nil
}

func (a *Monitor) GetHttpTypeParams() (MonitorHttpTypeParams, error) {
	var p MonitorHttpTypeParams
	if a.Params == "" {
		return p, nil
	}
	err := json.Unmarshal([]byte(a.Params), &p)
	return p, err
}

func (a *Monitor) SetTcpTypeParams(p MonitorTcpTypeParams) error {
	b, err := json.Marshal(p)
	if err != nil {
		return err
	}
	a.Params = string(b)
	return nil
}

func (a *Monitor) GetTcpTypeParams() (MonitorTcpTypeParams, error) {
	var p MonitorTcpTypeParams
	if a.Params == "" {
		return p, nil
	}
	err := json.Unmarshal([]byte(a.Params), &p)
	return p, err
}

func (a *Monitor) SetUdpTypeParams(p MonitorUdpTypeParams) error {
	b, err := json.Marshal(p)
	if err != nil {
		return err
	}
	a.Params = string(b)
	return nil
}

func (a *Monitor) GetUdpTypeParams() (MonitorUdpTypeParams, error) {
	var p MonitorUdpTypeParams
	if a.Params == "" {
		return p, nil
	}
	err := json.Unmarshal([]byte(a.Params), &p)
	return p, err
}

// 计划任务使用
func (a *Monitor) GetTaskID() string {
	return fmt.Sprintf("monitor_%d", a.ID)
}
