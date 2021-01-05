package http

import (
	"download/cmd"
	"fmt"
)

var TaskGroup map[string]*DownloadTask

func NewDownload(urls ...string) error {
	fmt.Println("HTTP connecting...")
	num := 0
	size := int64(0)
	for _, url := range urls {
		task := NewTask(url)
		err := task.GetResponseFile()
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", task.Name)
		task.PrintPbIdx = num
		TaskGroup[url] = task
		num++
		size += task.Size
	}
	fmt.Printf("共%d个任务，花费空间 %s, 是否继续？(y/n)", num, sizeString(size))
	var sure byte
	fmt.Scanf("%c", &sure)
	if sure == 'y' || sure == 'Y' {
		pb := cmd.PrintPb{}
		cmd.chpb = make(chan cmd.PrintInfo, size/1024/1024)
		go pb.bind()
		for _, url := range urls {
			go TaskGroup[url].Down()
		}
	}
	return nil
}
