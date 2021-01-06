package http

import (
	"download/cmd"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
)

type info struct {
	Name         string
	Url          string
	Header       map[string]string
	Size         int64
	SupportRange bool
	ChunkSize    int64
	ThreadNum    int
	Queue        [][2]int64
	FileName     string
}

type infoList struct {
	Tasks []info
}

var TaskSaveList infoList

func Save() error {
	j, err := json.Marshal(TaskSaveList)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("task.json", j, os.ModeAppend)
	if err != nil {
		return err
	}
	return nil
}

func Load() error {
	j := infoList{}
	b, err := ioutil.ReadFile("task.json")
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	taskList := []*DownloadTask{}
	for i, _ := range j.Tasks {
		task := info2task(&j.Tasks[i])
		taskList = append(taskList, task)
	}

	TaskGroup = make(map[string]*DownloadTask)

	for i, _ := range taskList {
		task := taskList[i]
		err := task.GetResponseFile()
		if err != nil {
			return err
		}
		fmt.Printf("task[%d] %s\n", i+1, task.Name)
		task.PrintPbIdx = i + 1
		TaskGroup[task.URL] = task
	}
	fmt.Printf("共检测到%d个未完成任务, 是否继续？(y/n)", len(taskList))
	var sure byte
	fmt.Scanf("%c", &sure)
	if sure == 'y' || sure == 'Y' {
		PB = cmd.NewPb(1000, len(taskList)+1)
		go PB.Bind()
		wg := &sync.WaitGroup{}
		wg.Add(len(taskList))
		for i, _ := range taskList {
			PB.PrintIdx(i + 1)
			go taskList[i].Down(wg)
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

func task2info(task *DownloadTask) *info {
	newinfo := &info{
		Name:         task.Name,
		Url:          task.URL,
		Header:       task.Header,
		Size:         task.Size,
		SupportRange: task.SupportRange,
		ChunkSize:    task.ChunkSize,
		ThreadNum:    task.ThreadNum,
		FileName:     task.FileName,
	}
	for len(task.Queue) > 0 {
		q := <-task.Queue
		newinfo.Queue = append(newinfo.Queue, q)
	}
	return newinfo
}

func info2task(info *info) *DownloadTask {
	task := &DownloadTask{
		Name:         info.Name,
		URL:          info.Url,
		Header:       info.Header,
		Status:       ALIVE,
		Size:         info.Size,
		SupportRange: info.SupportRange,
		ChunkSize:    info.ChunkSize,
		ThreadNum:    info.ThreadNum,
		SaveLoad:     true,
		FileName:     info.FileName,
	}
	task.Queue = make(chan [2]int64, len(info.Queue))
	for _, q := range info.Queue {
		task.Queue <- q
	}
	file, err := os.Open(info.FileName)
	if err != nil {
		log.Fatal(err)
	}
	task.File = file
	return task
}
