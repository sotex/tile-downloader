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
	"strings"
	"sync"
)

type (
	// TileLayer 瓦片图层基本信息
	TileLayer struct {
		MinZoom     int      // 默认0，该图层将显示的最小缩放级别(包括)
		MaxZoom     int      // 默认18，该图层将显示的最大缩放级别(包括)
		SubDomains  []string // tile服务的子域名
		ZoomOffset  int      // 默认0，瓦片URL中使用的 z = z + zoomOffset
		YReverse    bool     // 默认false，如果为true，则反转瓦片的Y轴编号
		ZoomReverse bool     // 默认false，如果设置为true，在tile url中使用的 z = maxZoom-z
	}
	TileDownloader struct {
		urlTemplate *template.Template
		layer       *TileLayer
		header      http.Header
	}
)

func NewTileLayer() *TileLayer {
	return &TileLayer{
		MinZoom:     0,
		MaxZoom:     18,
		SubDomains:  nil,
		ZoomOffset:  0,
		YReverse:    false,
		ZoomReverse: false,
	}
}

func NewTileDownloader(urlTemplate string, layer *TileLayer) *TileDownloader {
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
	td.header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:88.0) Gecko/20100101 Firefox/88.0")
	td.header.Set("Accept", "image/webp,*/*")
	td.header.Set("Accept-Language", "zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2")
	td.header.Set("Connection", "keep-alive")
	// td.header.Set("Referer", "http://map.geoq.cn/arcgis/rest/services/ChinaOnlineCommunity/MapServer?f=jsapi")
	return td
}

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

func (td *TileDownloader) Start(outDir string) {
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
						dirname := fmt.Sprintf("%s/Z%d/%d", outDir, z, y)
						filename := fmt.Sprintf("%s/Z%d/%d/%d.png", outDir, z, y, x)
						os.MkdirAll(dirname, os.ModePerm)
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
