package cmd

import (
	"fmt"
	"sync"
)

type PrintPb struct {
	mu       sync.Mutex
	currLine int
}

var chpb chan PrintInfo

type PrintInfo struct {
	idx int
	str string
}

func (pb *PrintPb) bind() {
	for {
		info, ok := <-chpb
		if ok {
			pb.mu.Lock()
			fmt.Printf("\033[%dA\033[%dB", pb.currLine, info.idx)
			pb.currLine = info.idx
			fmt.Printf(info.str)
			pb.mu.Unlock()
		} else {
			return
		}
	}
}
