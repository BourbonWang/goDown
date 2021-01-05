package http

import (
	"context"
	"fmt"
	"sync"
)

func (task *DownloadTask) Pause() {
	fmt.Println("pause")
	task.Status = PAUSE
}

func (task *DownloadTask) Continue() {
	task.ErrControl.mu.Lock()
	task.ErrControl.ErrNum = 0
	task.ErrControl.mu.Unlock()
	task.Status = ALIVE
}

func (task *DownloadTask) Finish() {
	if _, ok := TaskGroup[task.URL]; ok {
		delete(TaskGroup, task.URL)
	}
}

func (task *DownloadTask) Delete() {

}

type HTTPErr struct {
	ErrNum int
	mu     sync.Mutex
}

func (task *DownloadTask) BindHTTPErr(cancel context.CancelFunc) {
	for {
		if task.Status == ALIVE && task.ErrControl.ErrNum >= 16 {
			task.Pause()
		}
		//检测网络

		//取消任务，保存

		//cancel()
	}
}

func (task *DownloadTask) AddHTTPErr() {
	task.ErrControl.mu.Lock()
	task.ErrControl.ErrNum++
	task.ErrControl.mu.Unlock()
}
