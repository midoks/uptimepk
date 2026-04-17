package cmd

import (
	"fmt"

	"github.com/urfave/cli"

	"uptimepk/internal/app"
	"uptimepk/internal/conf"
	"uptimepk/internal/db"
	"uptimepk/internal/log"
)

var Web = cli.Command{
	Name:        "web",
	Usage:       "this command start web services",
	Description: `start web services`,
	Action:      runWeb,
	Flags: []cli.Flag{
		stringFlag("config, c", "", "custom configuration file path"),
	},
}

func runWeb(c *cli.Context) error {
	err := conf.InitConf(c.String("config"))
	if err != nil {
		fmt.Println("runWeb:", err)
		return err
	}

	log.Init()
	db.InitDb()

	//
	app.Run()
	return nil
}
