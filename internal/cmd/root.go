package cmd

import (
	"fmt"
	"github.com/urfave/cli"

	"uptimepk/internal/conf"
	"uptimepk/internal/db"
	"uptimepk/internal/log"
)

var Root = cli.Command{
	Name:        "root",
	Usage:       "this command modify root password",
	Description: `modify root password`,
	Action:      runRoot,
	Flags: []cli.Flag{
		stringFlag("config, c", "", "custom configuration file path"),
		stringFlag("password, p", "", "reset admin password"),
	},
}

// ./mgo_web root --password "admin"
func runRoot(c *cli.Context) error {
	conf.InitConf(c.String("config"))
	log.Init()
	db.InitDb()

	pwd := c.String("password")
	if pwd != "" {
		err := db.AdminUpdatePass(nil, 1, pwd)
		if err != nil {
			fmt.Println("update admin password fail!")
			return err
		}
		fmt.Println("update admin password success!")
	}
	return nil
}
