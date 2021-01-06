package cmd

import (
	"fmt"
	"sync"
)

type PrintPb struct {
	mu        sync.Mutex
	currLine  int
	ch        chan PrintInfo
	finishNum []int
	max       []int
}

type PrintInfo struct {
	idx int
	str string
}

func NewPb(size, num int) *PrintPb {
	pb := &PrintPb{}
	pb.ch = make(chan PrintInfo, size)
	pb.finishNum = make([]int, num)
	pb.max = make([]int, num)
	return pb
}

func (pb *PrintPb) Bind() {
	for {
		info, ok := <-pb.ch
		if ok {
			fmt.Printf("\033[%dA\033[%dB", pb.currLine, info.idx)
			pb.currLine = info.idx
			fmt.Printf(info.str)
		} else {
			return
		}
	}
}

func (pb *PrintPb) SetMaxNum(idx, n int) {
	pb.max[idx] = n
}

func (pb *PrintPb) Print(idx int, str string) {
	s := fmt.Sprintf("\rtask[%d]: finish chunk %4d / %d\t%s", idx, pb.finishNum[idx], pb.max[idx], str)
	pb.ch <- PrintInfo{idx, s}
}

func (pb *PrintPb) Add(idx int) {
	pb.mu.Lock()
	pb.finishNum[idx]++
	pb.mu.Unlock()
	pb.Print(idx, "下载中")
}

func (pb *PrintPb) PrintIdx(idx int) {
	s := fmt.Sprintf("\rtask[%d]:", idx)
	pb.ch <- PrintInfo{idx, s}
}

func (pb *PrintPb) PrintTail() {
	fmt.Printf("\033[%dA\033[%dB", pb.currLine, len(pb.finishNum))
}
