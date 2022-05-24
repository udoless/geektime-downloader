package application

import (
	"github.com/udoless/geektime-downloader/config"
	"github.com/udoless/geektime-downloader/service"
)

func getService() *service.Service {
	return config.Instance.ActiveUserService()
}
