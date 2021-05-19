## 简述

这是一个下载瓦片地图的小程序。之前做过很多次这样的事情，为了下载一点地图底图，写一个程序去下载瓦片，也就是临时用用。这次把地图瓦片下载这个功能，使用 Golang 进行一个简单的封装，以便后续使用的时候能够少做重复工作。

整个程序部分其实很短，主要都在 [tile_downloader.go](tile_downloader.go) 文件中了。这里没有封装为一个单独的包，需要使用的话直接把代码贴过去就是。

**待完善项**

- 添加 Grid 的封装，支持不同的切片规则，包括经纬度切片。
- 支持自动设置 Cookie，避免手动设置。
- 下载过程支持调用进度输出函数。
- 支持设置一个下载区域的矢量多边形，指定下载的范围。
- 支持输出瓦片到 SQLite 或者其他数据库，这个可以封装一个 WriteTile 的接口来实现。


## 使用说明

下载主要的参数都是参考 [Leaflet/TileLayer/options](https://leafletjs.com/reference-1.7.1.html#tilelayer-option) 的实现，在代码注释里面也有说明。

使用示例代码如下：

这是一个输出路径函数生成函数，是为了方便后面下载的时候使用的，也可以不使用它。

```go
func genGetOutPath(outdir string) func(z, x, y int) string {
	outdir = path.Clean(outdir)
	return func(z, x, y int) string {
		return fmt.Sprintf("%s/%d/%d/%d.png", outdir, z, x, y)
	}
}
```

**下载 ArcGIS 在线底图示例**

```go
	// 下载 ArcGIS 在线底图示例
	urltemplate := "https://map.geoq.cn/arcgis/rest/services/ChinaOnlineStreetGray/MapServer/tile/{z}/{y}/{x}"
	layer := NewTileLayer()
	layer.MinZoom = 0
	layer.MaxZoom = 2
	td := NewTileDownloader(urltemplate, layer, nil)
	td.Start(genGetOutPath("ChinaOnlineStreetGray"))
```

**下载天地图矢量底图示例**

下载天地图数据需要一个 `tk` 密钥，这个可以通过访问天地图的在线地图的时候获取。天地图的服务器会设置 Cookie ，这个也可以设置在头里面。

```go
    // 下载天地图矢量底图示例
	urltemplate := "https://{s}.tianditu.gov.cn/vec_w/wmts?SERVICE=WMTS&REQUEST=GetTile&VERSION=1.0.0" +
		"&LAYER=vec&STYLE=default&TILEMATRIXSET=w&FORMAT=tiles&tk=ef6151d9f0386f3b2a2fdf1d58fe9b32" +
		"&TILECOL={x}&TILEROW={y}&TILEMATRIX={z}"
	layer := NewTileLayer()
	layer.MinZoom = 0
	layer.MaxZoom = 2
	layer.SubDomains = []string{"t1", "t2", "t3", "t4"} // 设置子域名索引

    // 使用自定义的 HTTP 头，主要是后面的 Cookie 参数
	header := make(http.Header)
	header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:88.0) Gecko/20100101 Firefox/88.0")
	header.Set("Accept", "image/webp,*/*")
	header.Set("Accept-Language", "zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2")
	header.Set("Connection", "keep-alive")
	header.Set("Referer", "https://map.tianditu.gov.cn/")
	header.Set("Cookie", "HWWAFSESID=3cf10107e4f7f638a42f; HWWAFSESTIME=1603178985672")

	td := NewTileDownloader(urltemplate, layer, header)

	td.Start(genGetOutPath("vec_w"))
```