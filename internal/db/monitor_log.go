package db

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"uptimepk/internal/conf"
	"uptimepk/internal/model"
)

// getMonitorLogTableName 根据日期获取分表名
func getMonitorLogTableName(date time.Time) string {
	// 获取表前缀
	prefix := conf.Database.TablePrefix
	if prefix == "" {
		prefix = "uppk_"
	}

	// 格式化为 yyyymmdd 格式
	year, month, day := date.Date()
	dayStr := fmt.Sprintf("%04d%02d%02d", year, month, day)

	return prefix + "monitor_logs_" + dayStr
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

	// 计算日期范围（最近30天）
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	// 遍历日期范围，同时计算总记录数和收集数据
	var totalCount int64
	var resultList []model.MonitorLog
	remaining := size
	offset := (page - 1) * size

	currentDate := endDate // 从最近的日期开始
	for currentDate.After(startDate) || currentDate.Equal(startDate) {
		tableName := getMonitorLogTableName(currentDate)

		// 检查表格是否存在
		exists := GetDb().Migrator().HasTable(tableName)
		if !exists {
			currentDate = currentDate.AddDate(0, 0, -1)
			continue
		}

		// 计算当前表的数据量
		var count int64
		if err := GetDb().Table(tableName).Count(&count).Error; err != nil {
			currentDate = currentDate.AddDate(0, 0, -1)
			continue
		}

		// 累加总记录数
		totalCount += count

		// 如果还有数据需要获取
		if remaining > 0 {
			if offset > 0 {
				if int64(offset) >= count {
					offset -= int(count)
					currentDate = currentDate.AddDate(0, 0, -1)
					continue
				}

				var tableData []model.MonitorLog
				if err := GetDb().Table(tableName).Order(columnName("id") + " desc").Offset(offset).Limit(remaining).Find(&tableData).Error; err != nil {
					currentDate = currentDate.AddDate(0, 0, -1)
					continue
				}
				resultList = append(resultList, tableData...)
				remaining -= len(tableData)
				offset = 0
			} else {
				var tableData []model.MonitorLog
				if err := GetDb().Table(tableName).Order(columnName("id") + " desc").Limit(remaining).Find(&tableData).Error; err != nil {
					currentDate = currentDate.AddDate(0, 0, -1)
					continue
				}
				resultList = append(resultList, tableData...)
				remaining -= len(tableData)
			}
		}

		currentDate = currentDate.AddDate(0, 0, -1)
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

	// 计算日期范围（最近30天）
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	// 遍历日期范围，同时计算总记录数和收集数据
	var totalCount int64
	var resultList []model.MonitorLog
	remaining := size
	offset := (page - 1) * size

	currentDate := endDate // 从最近的日期开始
	for currentDate.After(startDate) || currentDate.Equal(startDate) {
		tableName := getMonitorLogTableName(currentDate)

		// 检查表格是否存在
		exists := GetDb().Migrator().HasTable(tableName)
		if !exists {
			currentDate = currentDate.AddDate(0, 0, -1)
			continue
		}

		// 计算当前表的数据量
		var count int64
		if err := GetDb().Table(tableName).Where("monitor_id = ?", monitor_id).Count(&count).Error; err != nil {
			currentDate = currentDate.AddDate(0, 0, -1)
			continue
		}

		// 累加总记录数
		totalCount += count

		// 如果还有数据需要获取
		if remaining > 0 {
			if offset > 0 {
				if int64(offset) >= count {
					offset -= int(count)
					currentDate = currentDate.AddDate(0, 0, -1)
					continue
				}

				var tableData []model.MonitorLog
				if err := GetDb().Table(tableName).Where("monitor_id = ?", monitor_id).Order(columnName("id") + " desc").Offset(offset).Limit(remaining).Find(&tableData).Error; err != nil {
					currentDate = currentDate.AddDate(0, 0, -1)
					continue
				}
				resultList = append(resultList, tableData...)
				remaining -= len(tableData)
				offset = 0
			} else {
				var tableData []model.MonitorLog
				if err := GetDb().Table(tableName).Where("monitor_id = ?", monitor_id).Order(columnName("id") + " desc").Limit(remaining).Find(&tableData).Error; err != nil {
					currentDate = currentDate.AddDate(0, 0, -1)
					continue
				}
				resultList = append(resultList, tableData...)
				remaining -= len(tableData)
			}
		}

		currentDate = currentDate.AddDate(0, 0, -1)
	}

	return resultList, totalCount, nil
}

func GetMonitorLogListByDate(monitor_id int64, day int64) ([]model.MonitorLog, error) {
	// 将 day 转换为 time.Time
	dayStr := strconv.FormatInt(day, 10)
	if len(dayStr) != 8 {
		return nil, errors.Errorf("invalid day format: %d", day)
	}

	year, _ := strconv.Atoi(dayStr[0:4])
	month, _ := strconv.Atoi(dayStr[4:6])
	dayInt, _ := strconv.Atoi(dayStr[6:8])

	targetDate := time.Date(year, time.Month(month), dayInt, 0, 0, 0, 0, time.Local)

	// 获取分表名
	tableName := getMonitorLogTableName(targetDate)

	var list []model.MonitorLog
	if err := GetDb().Table(tableName).Where("monitor_id = ?", monitor_id).Order(columnName("id") + " asc").Find(&list).Error; err != nil {
		return nil, errors.Wrapf(err, "failed get monitor log list by date")
	}
	return list, nil
}

func GetMonitorLatestLog(monitor_id int64) (*model.MonitorLog, error) {
	// 检查今天的表
	today := time.Now()
	tableName := getMonitorLogTableName(today)

	var log model.MonitorLog
	if err := GetDb().Table(tableName).Where("monitor_id = ?", monitor_id).Order(columnName("id") + " desc").First(&log).Error; err == nil {
		return &log, nil
	}

	// 如果今天没有数据，检查昨天的表
	yesterday := today.AddDate(0, 0, -1)
	tableName = getMonitorLogTableName(yesterday)
	if err := GetDb().Table(tableName).Where("monitor_id = ?", monitor_id).Order(columnName("id") + " desc").First(&log).Error; err == nil {
		return &log, nil
	}

	// 如果昨天也没有数据，返回错误
	return nil, errors.New("no monitor logs found")
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

	// 创建监控日志
	monitorLog := &model.MonitorLog{
		MonitorID:  strconv.FormatInt(monitorID, 10),
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
	tableName := getMonitorLogTableName(now)

	// 确保表存在
	if err := ensureMonitorLogTableExists(tableName); err != nil {
		return err
	}

	// 插入监控日志到指定分表
	return GetDb().Table(tableName).Create(monitorLog).Error
}

func MonitorLogDeleteByID(tx *gorm.DB, id int64, monitorID string) error {
	if tx == nil {
		tx = GetDb()
	}

	// 首先需要找到该日志记录所在的表
	// 检查今天的表
	today := time.Now()
	tableName := getMonitorLogTableName(today)

	var count int64
	if err := tx.Table(tableName).Where("id = ? AND monitor_id = ?", id, monitorID).Count(&count).Error; err == nil && count > 0 {
		var d model.MonitorLog
		return tx.Table(tableName).Where("id = ? AND monitor_id = ?", id, monitorID).Delete(&d).Error
	}

	// 检查昨天的表
	yesterday := today.AddDate(0, 0, -1)
	tableName = getMonitorLogTableName(yesterday)
	if err := tx.Table(tableName).Where("id = ? AND monitor_id = ?", id, monitorID).Count(&count).Error; err == nil && count > 0 {
		var d model.MonitorLog
		return tx.Table(tableName).Where("id = ? AND monitor_id = ?", id, monitorID).Delete(&d).Error
	}

	// 检查前天的表
	twoDaysAgo := today.AddDate(0, 0, -2)
	tableName = getMonitorLogTableName(twoDaysAgo)
	if err := tx.Table(tableName).Where("id = ? AND monitor_id = ?", id, monitorID).Count(&count).Error; err == nil && count > 0 {
		var d model.MonitorLog
		return tx.Table(tableName).Where("id = ? AND monitor_id = ?", id, monitorID).Delete(&d).Error
	}

	return errors.New("monitor log not found")
}

// MonitorLogDeleteByIDWithMonitorID 通过 monitorID 和 id 删除监控日志
func MonitorLogDeleteByIDWithMonitorID(tx *gorm.DB, id int64, monitorID int64) error {
	return MonitorLogDeleteByID(tx, id, strconv.FormatInt(monitorID, 10))
}

// ensureMonitorLogTableExists 确保监控日志表存在
func ensureMonitorLogTableExists(tableName string) error {
	// 检查表是否存在
	exists := GetDb().Migrator().HasTable(tableName)
	if exists {
		return nil
	}

	// 获取数据库类型
	dbType := conf.Database.Type

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
	return GetDb().Exec(createTableSQL).Error
}

// CreateMonitorLogTable 创建监控日志分表
func CreateMonitorLogTable() error {
	// 提前创建今天和明天的表
	today := time.Now()
	tomorrow := today.AddDate(0, 0, 1)

	// 创建今天的表
	todayTable := getMonitorLogTableName(today)
	if err := ensureMonitorLogTableExists(todayTable); err != nil {
		return errors.Wrapf(err, "failed create today's monitor log table")
	}

	// 创建明天的表
	tomorrowTable := getMonitorLogTableName(tomorrow)
	if err := ensureMonitorLogTableExists(tomorrowTable); err != nil {
		return errors.Wrapf(err, "failed create tomorrow's monitor log table")
	}

	return nil
}

// UpdateMonitorLog 更新监控日志
func UpdateMonitorLog(monitorID int64, id int64, updates map[string]interface{}) error {
	// 首先需要找到该日志记录所在的表
	// 检查今天的表
	today := time.Now()
	tableName := getMonitorLogTableName(today)

	var log model.MonitorLog
	if err := GetDb().Table(tableName).Where("id = ? AND monitor_id = ?", id, monitorID).First(&log).Error; err == nil {
		return GetDb().Table(tableName).Where("id = ? AND monitor_id = ?", id, monitorID).Updates(updates).Error
	}

	// 检查昨天的表
	yesterday := today.AddDate(0, 0, -1)
	tableName = getMonitorLogTableName(yesterday)
	if err := GetDb().Table(tableName).Where("id = ? AND monitor_id = ?", id, monitorID).First(&log).Error; err == nil {
		return GetDb().Table(tableName).Where("id = ? AND monitor_id = ?", id, monitorID).Updates(updates).Error
	}

	// 检查前天的表
	twoDaysAgo := today.AddDate(0, 0, -2)
	tableName = getMonitorLogTableName(twoDaysAgo)
	if err := GetDb().Table(tableName).Where("id = ? AND monitor_id = ?", id, monitorID).First(&log).Error; err == nil {
		return GetDb().Table(tableName).Where("id = ? AND monitor_id = ?", id, monitorID).Updates(updates).Error
	}

	return errors.New("monitor log not found")
}

// DeleteMonitorLogBeforeDays 删除指定天数之前的监控日志
func DeleteMonitorLogBeforeDays(days int) error {
	// 计算目标日期
	targetDate := time.Now().AddDate(0, 0, -days)

	// 计算日期范围
	endDate := targetDate.AddDate(0, 0, -1)  // 删除到 targetDate 的前一天
	startDate := endDate.AddDate(0, 0, -365) // 最多检查一年的数据

	// 遍历日期范围，删除指定日期之前的数据
	currentDate := startDate
	for currentDate.Before(endDate) || currentDate.Equal(endDate) {
		tableName := getMonitorLogTableName(currentDate)

		// 检查表格是否存在
		exists := GetDb().Migrator().HasTable(tableName)
		if !exists {
			currentDate = currentDate.AddDate(0, 0, 1)
			continue
		}

		// 直接删除整个表（因为整个表都是过期数据）
		if err := GetDb().Migrator().DropTable(tableName); err != nil {
			// 记录错误但继续处理其他表
			fmt.Printf("Error dropping monitor log table %s: %v\n", tableName, err)
		}

		currentDate = currentDate.AddDate(0, 0, 1)
	}

	return nil
}
