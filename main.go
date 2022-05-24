package main

import (
	"fmt"
	"log"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/udoless/geektime-downloader/cli/cmds"
	"github.com/udoless/geektime-downloader/config"
	"github.com/urfave/cli"
)

func init() {

	err := config.Instance.Init()
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	f, err := os.OpenFile("log.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)

	app := cmds.NewApp()
	app.Commands = []cli.Command{}
	app.Commands = append(app.Commands, cmds.NewLoginCommand()...)
	app.Commands = append(app.Commands, cmds.NewBuyCommand()...)
	app.Commands = append(app.Commands, cmds.NewCourseCommand()...)

	app.Action = cmds.DefaultAction

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}
