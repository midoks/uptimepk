package op

import (
	"fmt"

	"uptimepk/internal/conf"
	"uptimepk/internal/db"
)

// TableInfo 表信息结构体
type TableInfo struct {
	TableName string   `json:"table_name"`
	Size      float64  `json:"size"`    // 单位：MB
	Type      string   `json:"type"`    // 表类型：监控日志、系统日志、其他
	Actions   []string `json:"actions"` // 可执行的操作
}

// 临时结构体，用于扫描 SQL 结果
type tableInfoTemp struct {
	TableName string  `json:"table_name"`
	Size      float64 `json:"size"` // 单位：MB
}

// tableTypeInfo 表类型信息
type tableTypeInfo struct {
	Type    string
	Actions []string
}

// GetTableList 获取所有数据库表的名称和占用空间，并根据表名匹配规则添加类型和操作
func GetTableList() ([]TableInfo, error) {
	var tempTables []tableInfoTemp

	// 获取表前缀
	tablePrefix := conf.Database.TablePrefix
	if tablePrefix == "" {
		tablePrefix = "uppk_"
	}

	// 定义表类型映射表
	tableTypes := map[string]tableTypeInfo{
		tablePrefix + "monitor_logs_": {Type: "监控日志", Actions: []string{"delete", "clean"}},
		tablePrefix + "logs":          {Type: "系统日志", Actions: []string{"clean"}},
	}

	// 查询所有表的信息
	query := `
		SELECT 
			table_name, 
			ROUND((data_length + index_length) / 1024 / 1024, 2) as size 
		FROM 
			information_schema.tables 
		WHERE 
			table_schema = DATABASE() 
		ORDER BY 
			size DESC
	`

	if err := db.GetDb().Raw(query).Scan(&tempTables).Error; err != nil {
		return nil, fmt.Errorf("获取表信息失败: %v", err)
	}

	var tables []TableInfo
	for _, temp := range tempTables {
		tableName := temp.TableName
		tableInfo := TableInfo{
			TableName: tableName,
			Size:      temp.Size,
		}

		found := false

		for prefix, typeInfo := range tableTypes {
			if len(tableName) > len(prefix) && tableName[:len(prefix)] == prefix {
				tableInfo.Type = typeInfo.Type
				tableInfo.Actions = typeInfo.Actions
				found = true
				break
			}
		}

		if !found {
			tableInfo.Type = "系统"
			tableInfo.Actions = []string{"查看"}
		}

		tables = append(tables, tableInfo)
	}

	return tables, nil
}
