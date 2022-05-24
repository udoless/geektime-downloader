package application

import (
	"github.com/udoless/geektime-downloader/service"
)

//BuyProductAll product
func BuyProductAll() (*service.ProductAll, error) {
	return getService().BuyProductAll()
}

//BuyColumns all columns
func BuyColumns() (*service.Product, error) {
	all, err := BuyProductAll()
	return all.Columns, err
}

//BuyVideos all columns
func BuyVideos() (*service.Product, error) {
	all, err := BuyProductAll()
	return all.Videos, err
}
