package cmds

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/udoless/geektime-downloader/cli/version"
	"github.com/udoless/geektime-downloader/config"
	"github.com/udoless/geektime-downloader/utils"
	"github.com/urfave/cli"
)

var (
	_debug         bool
	_info          bool
	_stream        string
	_pdf           bool
	_mp3           bool
	appName        = filepath.Base(os.Args[0])
	configSaveFunc = func(c *cli.Context) error {
		err := config.Instance.Save()
		if err != nil {
			return errors.New("保存配置错误：" + err.Error())
		}
		return nil
	}
	authorizationFunc = func(c *cli.Context) error {
		if config.Instance.AcitveUID <= 0 {
			if len(config.Instance.Geektimes) > 0 {
				return config.ErrHasLoginedNotLogin
			}
			return config.ErrNotLogin
		}

		return nil
	}
)

//NewApp cli app
func NewApp() *cli.App {
	app := cli.NewApp()
	app.Name = appName
	app.Usage = "极客时间下载客户端"
	app.Version = fmt.Sprintf("%s", version.Version)
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf("%s version %s\n", app.Name, app.Version)
	}
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:        "debug,d",
			Usage:       "Turn on debug logs",
			Destination: &_debug,
		},
		cli.BoolFlag{
			Name:        "info, i",
			Usage:       "只输出视频信息",
			Destination: &_info,
		},
		cli.StringFlag{
			Name:        "stream, s",
			Usage:       "选择要下载的指定类型",
			Destination: &_stream,
		},
		cli.BoolFlag{
			Name:        "pdf, p",
			Usage:       "下载专栏PDF文档",
			Destination: &_pdf,
		},
		cli.BoolFlag{
			Name:        "mp3, m",
			Usage:       "下载专栏MP3音频",
			Destination: &_mp3,
		},
	}

	app.Before = func(c *cli.Context) error {
		if _debug {
			logrus.SetLevel(logrus.DebugLevel)
		}
		return nil
	}

	return app
}

//DefaultAction default action
func DefaultAction(c *cli.Context) error {
	_, err := config.Instance.ActiveUserService().User()
	if err != nil {
		return err
	}
	loginCh := make(chan struct{})
	loginRespCh := make(chan struct{})
	go func() {
		for {
			select {
			case <-loginCh:
				fmt.Println("Relogin....")
				err := loginByPhoneAndPassword(config.Instance.ActiveUser().PHONE, config.Instance.ActiveUser().PASSWORD)
				if err != nil {
					fmt.Printf("login failed: %s\n", err)
					continue
				}
				err = configSaveFunc(c)
				if err != nil {
					fmt.Printf("config save failed: %s\n", err)
					continue
				}
				err = config.Instance.Init()
				if err != nil {
					fmt.Printf("config reinit failed: %s\n", err)
					continue
				}
				loginRespCh <- struct{}{}
			}
			time.Sleep(3 * time.Second)
		}
	}()
	utils.LoginCh = loginCh
	utils.LoginRespCh = loginRespCh

	if len(c.Args()) == 0 {
		cli.ShowAppHelp(c)
		return nil
	}

	dlc := &NewDownloadCommand()[0]
	if dlc != nil {
		return dlc.Run(c)
	}

	return nil
}
