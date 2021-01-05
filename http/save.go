package http

var PATH = "task.json"

type info struct {
	name         string
	url          string
	header       map[string]string
	size         int64
	supportRange bool
	chunkSize    int64
	threadNum    int
	queue        [2]int64
}

type infoList struct {
	task []info
}

/**
func Save() error {
	list := infoList{}
	for _, task := range TaskGroup {
		list = append(list)
	}

	file, err := os.Create(PATH)
	if err != nil {
		return err
	}
	defer file.Close()

}

func Load() {

}

func task2info(task *DownloadTask) *info {
	newinfo := info{
		name:         task.Name,
		url:          task.URL,
		header:       task.Header,
		size:         task.Size,
		supportRange: task.SupportRange,
		chunkSize:    task.ChunkSize,
		threadNum:    task.ThreadNum,
	}

}
**/
