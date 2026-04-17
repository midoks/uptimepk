package op

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"uptimepk/internal/db"
	"uptimepk/internal/model"
	utils "uptimepk/internal/utils"
)

func InitAdmin(user string, pass string) error {
	data, err := db.GetAdminByID(1)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {

			salt := utils.RandString(16)
			admin := &model.Admin{
				Username:   user,
				Password:   model.TwoHashPwd(pass, salt),
				Salt:       salt,
				Status:     true,
				SuperAdmin: true,
				FullName:   "超级管理员",
			}

			admin.CreateTime = time.Now().Unix()
			admin.UpdateTime = time.Now().Unix()
			if err := db.CreateAdmin(nil, admin); err != nil {
				return err
			}
		}
	}

	fmt.Println("data:", data)
	return nil
}
