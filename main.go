package main

import (
	"log"
	"os"

	"github.com/urfave/cli"

	"uptimepk/internal/cmd"
	"uptimepk/internal/conf"
)

const (
	Version = "1.0.4"
	AppName = "uptimepk"
)

func init() {
	conf.App.Version = Version
	// if !(conf.App.RunMode == "prod") {
	// 	conf.App.Version = fmt.Sprintf("%s%d", Version, time.Now().Unix())
	// }
	conf.App.Name = AppName
}

func main() {
	app := cli.NewApp()
	app.Name = AppName
	app.Version = Version
	app.Usage = "uptimepk service"
	app.Commands = []cli.Command{
		cmd.Web,
		cmd.Root,
		cmd.Install,
		cmd.Uninstall,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("Failed to start application: %v", err)
	}
}
