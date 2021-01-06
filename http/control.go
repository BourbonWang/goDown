package http

import (
	"context"
	"net/http"
	"net/http/cookiejar"
	"os"
	"sync"
)

func (task *DownloadTask) Pause() {
	task.Status = PAUSE
}

func (task *DownloadTask) Continue() {
	task.ErrControl.mu.Lock()
	task.ErrControl.ErrNum = 0
	task.ErrControl.mu.Unlock()
	task.Status = ALIVE
}

func (task *DownloadTask) Finish(wg *sync.WaitGroup) {
	PB.Print(task.PrintPbIdx, "已完成")
	if _, ok := TaskGroup[task.URL]; ok {
		delete(TaskGroup, task.URL)
	}
	wg.Done()
}

type HTTPErr struct {
	ErrNum int
	mu     sync.Mutex
}

func (task *DownloadTask) BindHTTPErr(cancel context.CancelFunc) {
	for {
		if task.Status == ALIVE && task.ErrControl.ErrNum >= 16 {
			PB.Print(task.PrintPbIdx, "正在重连")
			task.Pause()
			//检测网络
			for i := 0; i < ReconnectNum; i++ {
				req, err := buildHTTPRequest("HEAD", task.URL, task.Header)
				if err != nil {
					continue
				}
				req.Header.Add("Range", "bytes=0-0")
				jar, _ := cookiejar.New(nil)
				httpClient := http.Client{Jar: jar}
				res, err := httpClient.Do(req)
				if err != nil {
					continue
				}
				if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusPartialContent {
					continue
				}
				res.Body.Close()

				PB.Print(task.PrintPbIdx, "恢复中")
				task.ErrControl.mu.Lock()
				task.ErrControl.ErrNum = 0
				task.ErrControl.mu.Unlock()
				task.Continue()
				break
			}
			//重连失败,保存或失败
			if task.Status == PAUSE {
				if SaveTempFile {
					task.save()
				} else {
					os.Remove("files/" + task.File.Name())
				}
				PB.Print(task.PrintPbIdx, "下载失败")
				cancel()
				return
			}
		}

	}
}

func (task *DownloadTask) AddHTTPErr() {
	task.ErrControl.mu.Lock()
	task.ErrControl.ErrNum++
	task.ErrControl.mu.Unlock()
}

func (task *DownloadTask) save() {
	info := task2info(task)
	TaskSaveList.Tasks = append(TaskSaveList.Tasks, *info)
}
