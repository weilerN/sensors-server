package In_memo_db

import (
	"fmt"
	"log"
	"strings"
)

const (
	MaxWorkersAmount = 8
)

type args struct {
	sw     *sensorWeekDB
	serial string
	day    int32
}

type WorkerPool interface {
	Run()              //dispatch the Worker Pool, and we need to execute this method before adding task to the Worker Pool
	AddTask(task args) //	AddTask(task func())
}

// GetInfoWP implementation of the WorkerPool interface
type GetInfoWP struct {
	maxWorker        int
	queuedTaskC      chan args
	workerResults    chan string
	totalTasksAmount int
}

func (w *GetInfoWP) Run() {
	for i := 0; i < w.maxWorker; i++ {
		go func() {
			for task := range w.queuedTaskC {
				w.workerResults <- task.sw.getInfoBySensorWeek(task.serial, task.day)
			}
		}()
	}
}

func (w *GetInfoWP) AddTask(task args) {
	w.queuedTaskC <- task
}

func NewWorkerPool(num int) *GetInfoWP {
	output := &GetInfoWP{}
	output.maxWorker = MaxWorkersAmount
	output.totalTasksAmount = num
	output.queuedTaskC = make(chan args, num)
	output.workerResults = make(chan string, num)
	return output
}

func GetInfoWorkerPool(sm *sensormap, day int32, resChan chan<- string) {
	//start workerpool
	wp := NewWorkerPool(sm.len())
	wp.Run()

	//add all sensorWeek as tasks to the queue
	for serial, sensorWeek := range sm.db {
		wp.AddTask(args{sensorWeek, serial, day})
	}

	//on another goroutine waiting to workerpool to finish
	go func() {
		var buffer strings.Builder
		i := 0
		for sensorRes := range wp.workerResults {
			if _, err := fmt.Fprintf(&buffer, "%v", sensorRes); err != nil {
				log.Println(err)
			}
			i++
			if i == wp.totalTasksAmount {
				break
			}
		}
		close(wp.workerResults)
		resChan <- buffer.String() //write output back to the caller's channel
	}()

}
