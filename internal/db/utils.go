package db

import (
	"fmt"

	"gorm.io/gorm"

	"uptimepk/internal/conf"
)

func columnName(name string) string {
	if conf.Database.Type == "postgres" {
		return fmt.Sprintf(`"%s"`, name)
	}
	return fmt.Sprintf("`%s`", name)
}

func addStorageOrder(db *gorm.DB) *gorm.DB {
	return db.Order(fmt.Sprintf("%s, %s", columnName("order"), columnName("id")))
}
