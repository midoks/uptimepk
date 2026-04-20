package form

type MonitorAdd struct {
	ID   int64  `form:"id"`
	Gid  int64  `form:"gid"` // gid
	Name string `form:"name"`
	Type string `form:"type"`

	Addr         string `form:"addr"`
	CheckContent string `form:"check_content"`
	UserAgent    string `form:"user_agent"`

	TcpHost string `form:"tcp_host"`
	TcpPort int    `form:"tcp_port"`

	UdpHost string `form:"udp_host"`
	UdpPort int    `form:"udp_port"`

	Interval   int    `form:"interval"`    // interval
	MaxRetries int    `form:"max_retries"` // max_retries
	Timeout    int    `form:"timeout"`     // timeout
	Status     bool   `form:"status"`      // status
	Mark       string `form:"mark"`        // mark
}

type MonitorGroupAdd struct {
	ID       int64  `form:"id"`
	Name     string `form:"name"`
	RealTime bool   `form:"real_time"`
	Status   bool   `form:"status"`
}

type MonitorCreateNode struct {
	Name string `form:"name"`
	Ip   string `form:"ip"`
}

type MonitorNodeDelete struct {
	NodeID int64 `form:"node_id"`
}

type MonitorNodeList struct {
	Page
	ClusterID int64 `form:"cluster_id"`
}
