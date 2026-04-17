package op

import (
	// "fmt"
	// "time"

	// "github.com/pkg/errors"
	// "gorm.io/gorm"

	"uptimepk/internal/db"
	// "uptimepk/internal/model"
	// utils "uptimepk/internal/utils"
)

func AddLog(uid int64, content string) error {
	return db.AddLog(nil, uid, content)
}
