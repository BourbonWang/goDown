package main

import (
	"download/http"
	"fmt"
	"log"
	"time"
)

func main() {
	URL := "https://wdl1.cache.wps.cn/wps/download/ep/Linux2019/9719/wps-office_11.1.0.9719_amd64.deb"
	URL2 := "https://ime.sogoucdn.com/dl/index/1608620566/sogoupinyin_2.4.0.2942_amd64.deb?st=F97CyjdG-0x4sghYTyCN0w&e=1609921899&fn=sogoupinyin_2.4.0.2942_amd64.deb"
	startTime := time.Now()
	err := http.NewDownload(URL, URL2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("download spend %.0f seconds.\n", time.Now().Sub(startTime).Seconds())
}
