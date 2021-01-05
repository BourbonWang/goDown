package main

import (
	"download/http"
	"fmt"
	"log"
	"time"
)

func main() {
	URL := "https://mirrors.tuna.tsinghua.edu.cn/linuxmint-cd/debian/lmde-4-cinnamon-64bit.iso"

	startTime := time.Now()
	err := http.NewDownload(URL)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("download spend %.0f seconds.\n", time.Now().Sub(startTime).Seconds())
}
