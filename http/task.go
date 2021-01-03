package http

import (
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
		"Connection":      "keep-alive",
	}

	ALIVE int = 1
	PAUSE int = 2

	TaskGroup map[string]*DownloadTask
)

type DownloadTask struct {
	Name         string
	URL          string
	Header       map[string]string
	Status       int
	Size         int64
	SupportRange bool
	ChunkSize    int64
	ThreadNum    int
	File         *os.File
	Wg           *sync.WaitGroup
	Queue        chan [2]int64

	ErrControl HTTPErr
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
	TaskGroup[url] = task
	return task
}

func (task *DownloadTask) AddHeader(key, value string) {
	task.Header[key] = value
}
