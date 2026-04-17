package form

type MonitorCreate struct {
	Name string `form:"name"`
}

type MonitorSubMenu struct {
	Number int64  `form:"number"`
	Name   string `form:"name"`
	Link   string `form:"link"`
}

type MonitorGroupAdd struct {
	ID        int64  `form:"id"`
	Name      string `form:"name"`
	Status    bool   `form:"status"`
	ClusterID int64  `form:"cluster_id"`
}

type MonitorCreateNode struct {
	Name      string `form:"name"`
	ClusterID int64  `form:"cluster_id"`
	Ip        string `form:"ip"`
}

type MonitorNodeDelete struct {
	NodeID int64 `form:"node_id"`
}

type MonitorNodeList struct {
	Page
	ClusterID int64 `form:"cluster_id"`
}
