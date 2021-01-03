package http

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

func (task *DownloadTask) Down() error {
	//建立连接
	err := task.getResponseFile()
	if err != nil {
		return err
	}
	//创建文件
	file, err := os.Create("files/" + task.Name)
	if err != nil {
		return err
	}
	defer file.Close()
	if err := file.Truncate(task.Size); err != nil {
		return nil
	}
	task.File = file

	fmt.Printf("%s would cost %s\n", task.Name, sizeString(task.Size))
	fmt.Printf("sure to download? (y / n)")
	var sure byte
	fmt.Scanf("%c", &sure)
	if sure == 'y' || sure == 'Y' {
		if task.SupportRange {
			chunkNum := int(task.Size/task.ChunkSize) + 1
			//下载队列
			task.Wg = &sync.WaitGroup{}
			task.Wg.Add(chunkNum)
			task.Queue = make(chan [2]int64, chunkNum)
			for i := 0; i < chunkNum; i++ {
				start := int64(i) * task.ChunkSize
				end := start + task.ChunkSize
				if i == chunkNum-1 {
					end = task.Size
				}
				task.Queue <- [2]int64{start, end - 1}
			}
			fmt.Printf("start download\n")
			go task.BindHTTPErr()
			//多线程下载
			for i := 0; i < task.ThreadNum; i++ {
				go task.mutithreadDown(i)
			}
			task.Wg.Wait()

		} else {
			task.Queue = make(chan [2]int64, 1)
			task.Queue <- [2]int64{0, task.Size - 1}
			task.Wg = &sync.WaitGroup{}
			task.Wg.Add(1)
			go task.mutithreadDown(0)
			task.Wg.Wait()
		}
		fmt.Println("download complete!")

	} else {
		err := os.Remove("files/" + task.Name)
		if err != nil {
			return err
		}
	}
	task.Finish()
	return nil
}

func (task *DownloadTask) getResponseFile() error {
	req, err := buildHTTPRequest("HEAD", task.URL, task.Header)
	if err != nil {
		return fmt.Errorf("ERROR: create http request failed\n")
	}
	//创建http连接
	fmt.Println("HTTP connecting...")
	req.Header.Add("Range", "bytes=0-0")
	jar, _ := cookiejar.New(nil)
	httpClient := http.Client{Jar: jar}
	res, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ERROR: can not get HTTP response from %s \n", task.URL)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("ERROR: response status error: %d\n", res.StatusCode)
	}
	//
	contentDisposition := res.Header.Get("Content-Disposition")
	//获取文件名
	if contentDisposition != "" {
		_, params, _ := mime.ParseMediaType(contentDisposition)
		filename := params["filename"]
		if filename != "" {
			task.Name = filename
		}
	} else {
		//从URL获取
		parse, err := url.Parse(req.URL.String())
		if err == nil {
			task.Name = lastSubString(parse.Path)
			if task.Name == "" {
				task.Name = "file" + time.Now().String()
			}
		}
	}
	//是否支持分段下载
	task.SupportRange = res.StatusCode == http.StatusPartialContent
	if task.SupportRange {
		contentRange := res.Header.Get("Content-Range")
		if contentRange != "" {
			total := lastSubString(contentRange)
			if total != "" && total != "*" {
				size, err := strconv.ParseInt(total, 10, 64)
				if err != nil {
					return fmt.Errorf("ERROR: set Content-Range error\n")
				}
				task.Size = size
			}
		}
	} else {
		contentLength := res.Header.Get("Content-Length")
		if contentLength != "" {
			size, err := strconv.ParseInt(contentLength, 10, 64)
			if err != nil {
				return fmt.Errorf("ERROR: set Content-Length error\n")
			}
			task.Size = size
		}
	}
	return nil
}

func (task *DownloadTask) mutithreadDown(index int) {
	for {
		for task.Status == PAUSE {
		} //等待暂停
		arr := <-task.Queue
		start, end := arr[0], arr[1]
		fmt.Printf("%d Downloading: %d to %d\n", index, start, end)
		task.downChunk(start, end)
	}
}

func (task *DownloadTask) downChunk(start int64, end int64) {
	httpReq, _ := buildHTTPRequest("GET", task.URL, task.Header)
	httpReq.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", start, end))
	jar, _ := cookiejar.New(nil)
	httpClient := http.Client{
		Jar:     jar,
		Timeout: 15 * time.Second,
	}
	httpRes, err := httpClient.Do(httpReq)
	if err != nil {
		fmt.Println(err)
		task.Queue <- [2]int64{start, end}
		task.AddHTTPErr()
		return
	}
	defer httpRes.Body.Close()

	buf := make([]byte, 8192)
	index := start
	for {
		n, err := httpRes.Body.Read(buf)
		if n > 0 {
			writeSize, err := task.File.WriteAt(buf[0:n], index)
			if err != nil {
				fmt.Println(err)
				task.Queue <- [2]int64{start, end}
				return
			}
			index += int64(writeSize)
		}
		if err != nil {
			if err != io.EOF {
				fmt.Println(err)
				task.Queue <- [2]int64{start, end}
				return
			}
			break
		}
	}
	task.Wg.Done()
}

//创建http请求头
func buildHTTPRequest(method string, url string, header map[string]string) (*http.Request, error) {
	httpReq, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range header {
		httpReq.Header.Add(k, v)
	}
	return httpReq, nil
}

func lastSubString(s string) string {
	index := strings.LastIndex(s, "/")
	if index != -1 {
		return s[index+1:]
	}
	return ""
}

func sizeString(size int64) string {
	if size < 1024*1024 {
		return fmt.Sprintf("%.2f Kb", float64(size)/1024)
	} else if size > 1024*1024*1024 {
		return fmt.Sprintf("%.2f Gb", float64(size)/1024/1024/1024)
	} else {
		return fmt.Sprintf("%.2f Mb", float64(size)/1024/1024)
	}
}
