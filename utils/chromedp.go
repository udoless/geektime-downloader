package utils

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/device"
)

var LoginCh chan struct{}
var LoginRespCh chan struct{}
var ErrRetry = errors.New("retry error")

//ColumnPrintToPDF print pdf
func ColumnPrintToPDF(aid int, filename string, cookies map[string]string) error {
	var buf []byte
	// create chrome instance
	ctx, cancel := chromedp.NewContext(
		context.Background(),
		chromedp.WithLogf(log.Printf),
	)
	defer cancel()

	// create a timeout
	ctx, cancel = context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	urlMap := make(map[string]struct{})
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch e := ev.(type) {
		case *network.EventRequestWillBeSent:
			if e.Type == "Image" && strings.Contains(e.Request.URL, "geekbang") && !strings.Contains(e.Request.URL, "aliyun") {
				//fmt.Printf("【network.EventRequestWillBeSent】%s, %s\n", e.Type, e.Request.URL)
				urlMap[e.Request.URL] = struct{}{}
				wg.Add(1)
			}
		case *network.EventResponseReceived:
			if e.Type == "Image" && strings.Contains(e.Response.URL, "geekbang") && !strings.Contains(e.Response.URL, "aliyun") {
				//fmt.Printf("【network.EventResponseReceived】%s, %s\n", e.Type, e.Response.URL)
				delete(urlMap, e.Response.URL)
				wg.Done()
			}
		}
	})
	imgLoadedCh := make(chan struct{})

	// url := `https://time.geekbang.org/column/article/169881`
	url := `https://time.geekbang.org/column/article/` + strconv.Itoa(aid)
	err := chromedp.Run(ctx,
		chromedp.Tasks{
			network.SetExtraHTTPHeaders(network.Headers(map[string]interface{}{
				"User-Agent": "Nimei_" + time.Now().String(),
			})),
			chromedp.Emulate(device.IPhone7),
			enableLifeCycleEvents(),
			setCookies(cookies),
			navigateAndWaitFor(url, "firstImagePaint"),
			// chromedp.WaitVisible("img", chromedp.ByQueryAll),
			chromedp.WaitReady("img", chromedp.ByQueryAll),
			chromedp.ActionFunc(func(ctx context.Context) error {
				s := `
				document.querySelector('.iconfont').parentElement.parentElement.style.display='none';
				document.querySelector('.Index_white_1gqaD>div.iconfont').style.display='none';
				var audioBar = document.querySelector('.audio-float-bar');
				if(audioBar){
					audioBar.style.display='none'
				}
				var bottom = document.querySelector('.sub-bottom-wrapper');
					if(bottom){
						bottom.style.display='none'
					}
					[...document.querySelectorAll('ul>li>div>div>div:nth-child(2)>span')].map(e=>e.click());
				`
				_, exp, err := runtime.Evaluate(s).Do(ctx)
				if err != nil {
					return err
				}

				if exp != nil {
					return exp
				}

				return nil
			}),
			chromedp.ActionFunc(func(ctx context.Context) error {
				s := `
					var divs = document.getElementsByTagName('div');
					for (var i = 0; i < divs.length; ++i){
						if(divs[i].innerText === "打开APP"){
							divs[i].parentNode.parentNode.style.display="none";
							break;
						}
					}
				`
				_, exp, err := runtime.Evaluate(s).Do(ctx)
				if err != nil {
					return err
				}

				if exp != nil {
					return exp
				}

				return nil
			}),
			//尽量确保图片都已加载
			/*
				chromedp.ActionFunc(func(ctx context.Context) error {
					s := `
						var maxTry = 3
						var loaded=false;
						var t = setInterval(function(){
							var img=document.getElementsByTagName('img');
							for(var i=0;i<img.length;i++){
									if(!img[i].complete){
										loaded = false
										break;
									};
							}
							if(loaded || --maxTry <=0){
								clearInterval(t)
							}
						}, 3000)
					`
					t1 := time.Now().Unix()
					_, exp, err := runtime.Evaluate(s).Do(ctx)
					fmt.Printf("Wait img complete seconds: %d\n", time.Now().Unix()-t1)
					if err != nil {
						return err
					}

					if exp != nil {
						return exp
					}

					return nil
				}), */
			chromedp.ActionFunc(func(ctx context.Context) error {
				go func() {
					time.Sleep(2 * time.Second)
					wg.Wait()
					imgLoadedCh <- struct{}{}
				}()
				select {
				case <-time.After(time.Second * 60):
					fmt.Printf("Image load timtout, not loaded: %s\n", (urlMap))
				case <-imgLoadedCh:
					fmt.Println("Image loaded")
				}

				var err error
				buf, _, err = page.PrintToPDF().WithPrintBackground(true).Do(ctx)
				return err
			}),
		},
	)

	if err != nil {
		header, body := getHtml(url)
		fmt.Printf("err:%s, url: %s, headers: %s, html: %s\n", err, url, header, body)
		LoginCh <- struct{}{}
		<-LoginRespCh

		return ErrRetry
	}
	return ioutil.WriteFile(filename, buf, 0644)
}

func getHtml(url string) (*network.Headers, string) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var result string
	resp, err := chromedp.RunResponse(ctx,
		chromedp.Navigate(url),
		chromedp.InnerHTML("html", &result, chromedp.ByQuery),
	)
	if err != nil {
		log.Fatal(err)
	}
	return &resp.Headers, result
}

func setCookies(cookies map[string]string) chromedp.ActionFunc {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		expr := cdp.TimeSinceEpoch(time.Now().Add(180 * 24 * time.Hour))

		for key, value := range cookies {
			err := network.SetCookie(key, value).WithExpires(&expr).WithDomain(".geekbang.org").WithHTTPOnly(true).Do(ctx)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func enableLifeCycleEvents() chromedp.ActionFunc {
	return func(ctx context.Context) error {
		err := page.Enable().Do(ctx)
		if err != nil {
			return err
		}

		return page.SetLifecycleEventsEnabled(true).Do(ctx)
	}
}

func navigateAndWaitFor(url string, eventName string) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		_, _, _, err := page.Navigate(url).Do(ctx)
		if err != nil {
			return err
		}

		return waitFor(ctx, eventName)
	}
}

// waitFor blocks until eventName is received.
// Examples of events you can wait for:
//     init, DOMContentLoaded, firstPaint,
//     firstContentfulPaint, firstImagePaint,
//     firstMeaningfulPaintCandidate,
//     load, networkAlmostIdle, firstMeaningfulPaint, networkIdle
//
// This is not super reliable, I've already found incidental cases where
// networkIdle was sent before load. It's probably smart to see how
// puppeteer implements this exactly.
func waitFor(ctx context.Context, eventName string) error {
	ch := make(chan struct{})
	cctx, cancel := context.WithCancel(ctx)
	chromedp.ListenTarget(cctx, func(ev interface{}) {
		switch e := ev.(type) {
		case *page.EventLifecycleEvent:
			if e.Name == eventName {
				fmt.Println(e.Name)
				cancel()
				close(ch)
			} else {
				fmt.Print(e.Name + " -> ")
			}
		}
	})

	select {
	case <-ch:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
