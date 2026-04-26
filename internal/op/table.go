package op

import (
	"fmt"

	"uptimepk/internal/db"
)

// TableInfo 表信息结构体
type TableInfo struct {
	TableName string  `json:"table_name"`
	Size      float64 `json:"size"` // 单位：MB
	Use       string  `json:"use"`
}

// GetTableList 获取所有数据库表的名称和占用空间
func GetTableList() ([]TableInfo, error) {
	var tables []TableInfo

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

	if err := db.GetDb().Raw(query).Scan(&tables).Error; err != nil {
		return nil, fmt.Errorf("获取表信息失败: %v", err)
	}

	return tables, nil
}
