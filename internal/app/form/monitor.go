package form

type MonitorAdd struct {
	ID   int64  `form:"id"`
	Gid  int64  `form:"gid"` // gid
	Name string `form:"name"`
	Type string `form:"type"`

	Addr                 string `form:"addr"`
	CheckContent         string `form:"check_content"`
	UserAgent            string `form:"user_agent"`
	RelateMonitorGroupID int64  `form:"relate_monitor_group_id"`

	TcpHost string `form:"tcp_host"`
	TcpPort int    `form:"tcp_port"`

	UdpHost string `form:"udp_host"`
	UdpPort int    `form:"udp_port"`

	Interval     int    `form:"interval"`      // interval
	IntervalType string `form:"interval_type"` // interval_type
	MaxRetries   int    `form:"max_retries"`   // max_retries
	Timeout      int    `form:"timeout"`       // timeout
	Status       bool   `form:"status"`        // status
	Mark         string `form:"mark"`          // mark
}

type MonitorGroupAdd struct {
	ID       int64  `form:"id"`
	Name     string `form:"name"`
	RealTime bool   `form:"real_time"`
	Status   bool   `form:"status"`
}

type MonitorList struct {
	Page
	Gid int64  `form:"gid"`
	Key string `form:"key"`
}

type MonitorLogList struct {
	Page

	Times     string `form:"times"`
	Type      string `form:"type"`
	Key       string `form:"key"`
	StartTime string `form:"start_time"`
	EndTime   string `form:"end_time"`
}
