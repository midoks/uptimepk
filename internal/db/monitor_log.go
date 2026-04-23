package db

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"uptimepk/internal/conf"
	"uptimepk/internal/model"
)

// getMonitorLogTableName 根据 MonitorID 获取分表名
func getMonitorLogTableName(monitorID string) string {
	// 获取表前缀
	prefix := conf.Database.TablePrefix
	if prefix == "" {
		prefix = "uppk_"
	}

	// 使用 MD5 计算哈希值
	hash := md5.Sum([]byte(monitorID))
	hashStr := hex.EncodeToString(hash[:])

	// 取哈希值的前两位作为索引，转换为整数
	index := 0
	for i := 0; i < 2; i++ {
		if hashStr[i] >= '0' && hashStr[i] <= '9' {
			index = index*16 + int(hashStr[i]-'0')
		} else if hashStr[i] >= 'a' && hashStr[i] <= 'f' {
			index = index*16 + int(hashStr[i]-'a'+10)
		}
	}

	// 对 128 取模，得到 0-127 的索引
	index = index % 128

	return prefix + "monitor_logs_" + strconv.Itoa(index)
}

func GetMonitorLogList(page, size int) ([]model.MonitorLog, int64, error) {
	// 确保 page 至少为 1
	if page <= 0 {
		page = 1
	}
	// 确保 size 至少为 1
	if size <= 0 {
		size = 10
	}

	// 获取表前缀
	prefix := conf.Database.TablePrefix
	if prefix == "" {
		prefix = "uppk_"
	}

	// 遍历所有 128 个分表，计算总记录数
	var totalCount int64
	for i := 0; i < 128; i++ {
		tableName := prefix + "monitor_logs_" + strconv.Itoa(i)
		var count int64
		if err := GetDb().Table(tableName).Model(&model.MonitorLog{}).Count(&count).Error; err != nil {
			// 忽略表不存在的错误
			continue
		}
		totalCount += count
	}

	// 直接在数据库层面进行分页查询，不加载所有数据到内存
	var resultList []model.MonitorLog
	offset := (page - 1) * size
	remaining := size
	tableIndex := 0

	// 遍历分表收集数据直到填满一页
	for tableIndex < 128 && remaining > 0 {
		tableName := prefix + "monitor_logs_" + strconv.Itoa(tableIndex)
		var tableData []model.MonitorLog

		// 尝试从当前表获取数据
		query := GetDb().Table(tableName).Order(columnName("id") + " desc")

		if offset > 0 {
			// 跳过前面表的数据
			var count int64
			if err := GetDb().Table(tableName).Count(&count).Error; err != nil {
				tableIndex++
				continue
			}

			if int64(offset) >= count {
				offset -= int(count)
				tableIndex++
				continue
			}

			// 从当前表获取部分数据
			if err := query.Offset(offset).Limit(remaining).Find(&tableData).Error; err != nil {
				tableIndex++
				continue
			}
			offset = 0
		} else {
			// 直接从当前表获取剩余所需数据
			if err := query.Limit(remaining).Find(&tableData).Error; err != nil {
				tableIndex++
				continue
			}
		}

		resultList = append(resultList, tableData...)
		remaining -= len(tableData)
		tableIndex++
	}

	return resultList, totalCount, nil
}

func GetMonitorLogListByMonitorID(monitor_id int64, page, size int) ([]model.MonitorLog, int64, error) {
	// 确保 page 至少为 1
	if page <= 0 {
		page = 1
	}
	// 确保 size 至少为 1
	if size <= 0 {
		size = 10
	}

	// 转换 monitorID 为字符串
	monitorIDStr := strconv.FormatInt(monitor_id, 10)

	// 获取分表名
	tableName := getMonitorLogTableName(monitorIDStr)

	mmlog := GetDb().Table(tableName).Model(&model.MonitorLog{})
	var count int64
	if err := mmlog.Where("monitor_id = ? ", monitor_id).Count(&count).Error; err != nil {
		return nil, 0, errors.Wrapf(err, "failed get monitor log count")
	}

	var list []model.MonitorLog
	if err := GetDb().Table(tableName).Where("monitor_id = ? ", monitor_id).Order(columnName("id") + " desc").Offset((page - 1) * size).Limit(size).Find(&list).Error; err != nil {
		return nil, 0, errors.Wrapf(err, "failed get monitor log list")
	}
	return list, count, nil
}

func GetMonitorLogListByDate(monitor_id int64, day int64) ([]model.MonitorLog, error) {
	// 转换 monitorID 为字符串
	monitorIDStr := strconv.FormatInt(monitor_id, 10)

	// 获取分表名
	tableName := getMonitorLogTableName(monitorIDStr)

	var list []model.MonitorLog
	if err := GetDb().Table(tableName).Where("monitor_id = ? AND day = ?", monitor_id, day).Order(columnName("id") + " asc").Find(&list).Error; err != nil {
		return nil, errors.Wrapf(err, "failed get monitor log list by date")
	}
	return list, nil
}

func GetMonitorLatestLog(monitor_id int64) (*model.MonitorLog, error) {
	// 转换 monitorID 为字符串
	monitorIDStr := strconv.FormatInt(monitor_id, 10)

	// 获取分表名
	tableName := getMonitorLogTableName(monitorIDStr)

	var log model.MonitorLog
	if err := GetDb().Table(tableName).Where("monitor_id = ? ", monitor_id).Order(columnName("id") + " desc").First(&log).Error; err != nil {
		return nil, err
	}
	return &log, nil
}

// CreateMonitorLog 创建并插入监控日志
func CreateMonitorLog(monitorID int64, isValid bool, size int, speed float64, errorMsg string, maxRetries int) error {
	// 获取当前时间
	now := time.Now()
	year, month, day := now.Date()
	hour, minute, _ := now.Clock()
	timestamp := now.Unix()

	// 计算 yyyymmdd 格式的日期
	dayInt := year*10000 + int(month)*100 + day

	// speed 保留2位小数
	speed = math.Round(speed*100) / 100

	// 转换 monitorID 为字符串
	monitorIDStr := strconv.FormatInt(monitorID, 10)

	// 创建监控日志
	monitorLog := &model.MonitorLog{
		MonitorID:  monitorIDStr,
		Day:        int64(dayInt),
		Hour:       int64(hour),
		Minute:     minute,
		IsValid:    isValid,
		Size:       int64(size),
		Speed:      speed,
		ErrorMsg:   errorMsg,
		MaxRetries: maxRetries,
		CreateTime: timestamp,
	}

	// 获取分表名
	tableName := getMonitorLogTableName(monitorIDStr)

	// 插入监控日志到指定分表
	return GetDb().Table(tableName).Create(monitorLog).Error
}

func MonitorLogDeleteByID(tx *gorm.DB, id int64, monitorID string) error {
	if tx == nil {
		tx = GetDb()
	}

	// 获取分表名
	tableName := getMonitorLogTableName(monitorID)

	var d model.MonitorLog
	return tx.Table(tableName).Where("id = ?", id).Delete(&d).Error
}

// MonitorLogDeleteByIDWithMonitorID 通过 monitorID 和 id 删除监控日志
func MonitorLogDeleteByIDWithMonitorID(tx *gorm.DB, id int64, monitorID int64) error {
	return MonitorLogDeleteByID(tx, id, strconv.FormatInt(monitorID, 10))
}

// CreateMonitorLogTable 创建监控日志分表
func CreateMonitorLogTable() error {
	// 获取表前缀
	prefix := conf.Database.TablePrefix
	if prefix == "" {
		prefix = "uppk_"
	}

	// 获取数据库类型
	dbType := conf.Database.Type

	// 为每个分表创建表结构
	for i := 0; i < 128; i++ {
		tableName := prefix + "monitor_logs_" + strconv.Itoa(i)
		// 检查表是否存在
		exists := GetDb().Migrator().HasTable(tableName)
		if !exists {
			var createTableSQL string
			switch dbType {
			case "sqlite3":
				createTableSQL = fmt.Sprintf(`
				CREATE TABLE IF NOT EXISTS %s (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					monitor_id TEXT NOT NULL,
					day INTEGER NOT NULL,
					hour INTEGER NOT NULL,
					minute INTEGER NOT NULL,
					is_valid INTEGER NOT NULL,
					size INTEGER NOT NULL,
					speed REAL NOT NULL,
					error_msg TEXT,
					max_retries INTEGER NOT NULL,
					create_time INTEGER NOT NULL
				);
				`, tableName)
			case "mysql":
				createTableSQL = fmt.Sprintf(`
				CREATE TABLE IF NOT EXISTS %s (
					id BIGINT PRIMARY KEY AUTO_INCREMENT,
					monitor_id VARCHAR(255) NOT NULL,
					day BIGINT NOT NULL,
					hour BIGINT NOT NULL,
					minute INT NOT NULL,
					is_valid BOOLEAN NOT NULL,
					size BIGINT NOT NULL,
					speed DOUBLE NOT NULL,
					error_msg TEXT,
					max_retries INT NOT NULL,
					create_time BIGINT NOT NULL,
					INDEX idx_monitor_id (monitor_id),
					INDEX idx_day (day),
					INDEX idx_create_time (create_time)
				) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
				`, tableName)
			case "postgres":
				createTableSQL = fmt.Sprintf(`
				CREATE TABLE IF NOT EXISTS %s (
					id BIGSERIAL PRIMARY KEY,
					monitor_id VARCHAR(255) NOT NULL,
					day BIGINT NOT NULL,
					hour BIGINT NOT NULL,
					minute INT NOT NULL,
					is_valid BOOLEAN NOT NULL,
					size BIGINT NOT NULL,
					speed DOUBLE PRECISION NOT NULL,
					error_msg TEXT,
					max_retries INT NOT NULL,
					create_time BIGINT NOT NULL
				);
				CREATE INDEX IF NOT EXISTS idx_%s_monitor_id ON %s (monitor_id);
				CREATE INDEX IF NOT EXISTS idx_%s_day ON %s (day);
				CREATE INDEX IF NOT EXISTS idx_%s_create_time ON %s (create_time);
				`, tableName, tableName, tableName, tableName, tableName, tableName, tableName)
			default:
				return errors.Errorf("unsupported database type: %s", dbType)
			}

			// 创建表
			if err := GetDb().Exec(createTableSQL).Error; err != nil {
				return errors.Wrapf(err, "failed create monitor log table: %s", tableName)
			}
		}
	}
	return nil
}

// UpdateMonitorLog 更新监控日志
func UpdateMonitorLog(monitorID int64, id int64, updates map[string]interface{}) error {
	// 转换 monitorID 为字符串
	monitorIDStr := strconv.FormatInt(monitorID, 10)

	// 获取分表名
	tableName := getMonitorLogTableName(monitorIDStr)

	return GetDb().Table(tableName).Where("id = ? AND monitor_id = ?", id, monitorID).Updates(updates).Error
}
