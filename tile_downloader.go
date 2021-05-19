/*******************************************************************************
* @文件描述: 瓦片下载器封装
* @创建人员: solym  ymwh@foxmail.com
* @创建时间: 2021年05月18日14:33:55
* @更新人员: solym  ymwh@foxmail.com
* @更新时间: 2021年05月18日14:33:55
* @版权声明: None
* @文件名称: tile_downloader.go
* @修改历史:
********************************************************************************/

package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
)

type (
	// TileLayer 瓦片图层基本信息.
	//   这里的参数和 Leaflet 里面瓦片图层的 Options 是类似的.
	TileLayer struct {
		MinZoom     int      // 默认0，该图层将显示的最小缩放级别(包括)
		MaxZoom     int      // 默认18，该图层将显示的最大缩放级别(包括)
		SubDomains  []string // tile服务的子域名
		ZoomOffset  int      // 默认0，瓦片URL中使用的 z = z + zoomOffset
		YReverse    bool     // 默认false，如果为true，则反转瓦片的Y轴编号
		ZoomReverse bool     // 默认false，如果设置为true，在tile url中使用的 z = maxZoom-z
	}

	// TileDownloader 瓦片下载器.
	TileDownloader struct {
		urlTemplate *template.Template // 瓦片地址的 URL 模板，和 Leaflet 里面是一样的规则
		layer       *TileLayer         // 瓦片图层信息
		header      http.Header        // 下载过程时候的 HTTP 头信息，以便添加自定义的 HTTP 头等
	}
)

// NewTileLayer 返回一个设置默认参数的 TileLayer 实例
func NewTileLayer() *TileLayer {
	return &TileLayer{
		MaxZoom: 18,
	}
}

// NewTileDownloader 返回一个瓦片下载器实例
//   urlTemplate  瓦片的下载地址模板
//   layer        瓦片图层的基本信息
func NewTileDownloader(urlTemplate string, layer *TileLayer, header http.Header) *TileDownloader {
	urlTemplate = strings.ReplaceAll(urlTemplate, "{s}", "{{.S}}")
	urlTemplate = strings.ReplaceAll(urlTemplate, "{x}", "{{.X}}")
	urlTemplate = strings.ReplaceAll(urlTemplate, "{y}", "{{.Y}}")
	urlTemplate = strings.ReplaceAll(urlTemplate, "{z}", "{{.Z}}")

	ut, err := template.New("url").Parse(urlTemplate)
	if err != nil {
		return nil
	}

	td := &TileDownloader{
		urlTemplate: ut,
		layer:       layer,
		header:      make(http.Header),
	}
	td.header = header
	if td.header == nil {
		td.header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:88.0) Gecko/20100101 Firefox/88.0")
		td.header.Set("Accept", "image/webp,*/*")
		td.header.Set("Accept-Language", "zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2")
		td.header.Set("Connection", "keep-alive")
	}
	return td
}

// 下载一个瓦片数据
func (td *TileDownloader) downloadTile(z, x, y int) (data []byte, err error) {
	var tileIndex = struct {
		X int
		Y int
		Z int
		S string
	}{
		X: x,
		Y: y,
		Z: z + td.layer.ZoomOffset,
	}
	if td.layer.YReverse {
		tileIndex.Y = -tileIndex.Y
	}
	if len(td.layer.SubDomains) > 0 {
		tileIndex.S = td.layer.SubDomains[tileIndex.X%len(td.layer.SubDomains)]
	}

	var sb strings.Builder
	if err = td.urlTemplate.Execute(&sb, tileIndex); err != nil {
		return
	}

	req, err := http.NewRequest("GET", sb.String(), nil)
	if err != nil {
		return
	}
	req.Header = td.header

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

// Start 启动下载，会阻塞直到全部下载任务结束
//   fnGetOutPath 是用于获取瓦片输出路径的函数，可以自己根据需要进行编写
//   下载过程中会同时进行四个任务的下载，等全部任务结束后才会退出
func (td *TileDownloader) Start(getOutPath func(z, x, y int) string) {
	ch := make(chan int, 4)
	var wg sync.WaitGroup
	for z := td.layer.MinZoom; z <= td.layer.MaxZoom; z++ {
		m := 0x1 << z
		for y := 0; y < m; y++ {
			for x := 0; x < m; x++ {
				wg.Add(1)
				go func(x, y, z int) {
					ch <- 1
					data, err := td.downloadTile(z, x, y)
					if err == nil {
						filename := getOutPath(z, x, y)
						os.MkdirAll(path.Dir(filename), os.ModePerm)
						ioutil.WriteFile(filename, data, 0755)
					} else {
						fmt.Println("错误:", err)
					}
					<-ch
					wg.Done()
				}(x, y, z)
			}
		}
	}
	wg.Wait()
}
