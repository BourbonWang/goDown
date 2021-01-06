package http

import (
	"download/cmd"
	"fmt"
	"log"
	"os"
	"sync"
)

var (
	DefaultHeader = map[string]string{
		"Accept": "text/html,application/xhtml+xml,application/xml;q=0.9," +
			"image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
		"Accept-Encoding": "gzip, deflate, br",
		"Accept-Language": "zh-CN,zh;q=0.9",
		"User-Agent":      "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.111 Safari/537.36",
		//"Connection":      "keep-alive",
	}
	ReconnectNum = 10
	SaveTempFile = true
	BasePath     = "files/"

	ALIVE int = 1
	PAUSE int = 2
)

var TaskGroup map[string]*DownloadTask
var PB *cmd.PrintPb

type DownloadTask struct {
	Name         string
	URL          string
	Header       map[string]string
	Status       int
	Size         int64
	SupportRange bool
	ChunkSize    int64
	ThreadNum    int
	FileName     string
	File         *os.File
	Queue        chan [2]int64
	PrintPbIdx   int
	ErrControl   HTTPErr
	SaveLoad     bool
}

func NewDownload(urls ...string) error {
	TaskGroup = make(map[string]*DownloadTask)

	fmt.Println("HTTP connecting...")
	size := int64(0)
	for i, url := range urls {
		task := NewTask(url)
		err := task.GetResponseFile()
		if err != nil {
			return err
		}
		fmt.Printf("task[%d] %s\n", i+1, task.Name)
		task.PrintPbIdx = i + 1
		TaskGroup[url] = task
		size += task.Size
	}
	fmt.Printf("共%d个任务，花费空间 %s, 是否继续？(y/n)", len(urls), sizeString(size))
	var sure byte
	fmt.Scanf("%c", &sure)
	if sure == 'y' || sure == 'Y' {
		PB = cmd.NewPb(int(size/1024/1024), len(urls)+1)
		go PB.Bind()
		wg := &sync.WaitGroup{}
		wg.Add(len(urls))
		for i, url := range urls {
			PB.PrintIdx(i + 1)
			err := TaskGroup[url].CreateFile()
			if err != nil {
				return err
			}
			go TaskGroup[url].Down(wg)
		}
		wg.Wait()
		PB.PrintTail()
		fmt.Println()
		if SaveTempFile {
			err := Save()
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("已保存临时文件")
			}
		}
		fmt.Printf("下载结束")
	}
	return nil
}

func NewTask(url string) *DownloadTask {
	if _, ok := TaskGroup[url]; ok {
		log.Fatal("task is already exist.\n")
		return nil
	}

	task := &DownloadTask{
		URL:       url,
		Header:    DefaultHeader,
		ChunkSize: int64(1024 * 1024),
		ThreadNum: 16,
		Status:    ALIVE,
	}

	return task
}

func (task *DownloadTask) AddHeader(key, value string) {
	task.Header[key] = value
}
