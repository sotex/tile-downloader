/*******************************************************************************
* @文件描述: 使用示例代码
* @创建人员: solym  ymwh@foxmail.com
* @创建时间: 2021年05月18日14:05:48
* @更新人员: solym  ymwh@foxmail.com
* @更新时间: 2021年05月19日13:03:32
* @版权声明: None
* @文件名称: main.go
* @修改历史:
********************************************************************************/

package main

import (
	"fmt"
	"net/http"
	"path"
)

func main() {

	// 下载 ArcGIS 在线底图示例
	if false {
		urltemplate := "http://map.geoq.cn/arcgis/rest/services/ChinaOnlineCommunity/MapServer/tile/{z}/{y}/{x}"
		// urltemplate := "https://map.geoq.cn/arcgis/rest/services/ChinaOnlineStreetGray/MapServer/tile/{z}/{y}/{x}"
		layer := NewTileLayer()
		layer.MinZoom = 0
		layer.MaxZoom = 2
		td := NewTileDownloader(urltemplate, layer, nil)
		td.Start(genGetOutPath("ChinaOnlineStreetGray"))
	}

	// 下载天地图矢量底图示例
	if false {
		urltemplate := "https://t{s}.tianditu.gov.cn/vec_w/wmts?SERVICE=WMTS&REQUEST=GetTile&VERSION=1.0.0" +
			"&LAYER=vec&STYLE=default&TILEMATRIXSET=w&FORMAT=tiles&tk=ef6151d9f0386f3b2a2fdf1d58fe9b32" +
			"&TILECOL={x}&TILEROW={y}&TILEMATRIX={z}"
		layer := NewTileLayer()
		layer.MinZoom = 0
		layer.MaxZoom = 2
		layer.SubDomains = []string{"1", "2", "3", "4"}

		header := make(http.Header)
		header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:88.0) Gecko/20100101 Firefox/88.0")
		header.Set("Accept", "image/webp,*/*")
		header.Set("Accept-Language", "zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2")
		header.Set("Connection", "keep-alive")
		header.Set("Referer", "https://map.tianditu.gov.cn/")
		header.Set("Cookie", "HWWAFSESID=3cf10107e4f7f638a42f; HWWAFSESTIME=1603178985672")

		td := NewTileDownloader(urltemplate, layer, header)

		td.Start(genGetOutPath("vec_w"))
	}
}

func genGetOutPath(outdir string) func(z, x, y int) string {
	outdir = path.Clean(outdir)
	return func(z, x, y int) string {
		return fmt.Sprintf("%s/%d/%d/%d.png", outdir, z, x, y)
	}
}
