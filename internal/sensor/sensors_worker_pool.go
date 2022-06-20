package sensor

import (
	grpc_db "SensorServer/pkg/grpc_db"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"math/rand"
	"time"
)

const (
	MaxWorkersAmount = 8
)

type args struct {
	Conn *grpc_db.SensorStreamClient
}

type WorkerPool interface {
	Run()              //dispatch the Worker Pool, and we need to execute this method before adding task to the Worker Pool
	AddTask(task args) //	AddTask(task func())
}

// SensorWorkerPool implementation of the WorkerPool interface
type SensorWorkerPool struct {
	maxWorker          int
	queuedTaskC        chan args
	totalSensorsAmount int
	ctx                context.Context
}

func randMeasure() int32 {
	r1 := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r1.Int31() % 30
}

func randSerial(workerNum int) string {
	r1 := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("sensor_%d", r1.Intn(workerNum))
}

func (w *SensorWorkerPool) Run() {
	for i := 0; i < w.maxWorker; i++ {
		go func() {
			for task := range w.queuedTaskC {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				_, err := (*task.Conn).SensorMeasure(ctx, &grpc_db.Measure{
					M:      randMeasure(),
					Serial: randSerial(w.totalSensorsAmount),
				})
				cancel()
				if err != nil {
					log.Fatalf("%v", err)
				}
				//return the connection back to the chan
				w.queuedTaskC <- task
			}
		}()
	}
}

func (w *SensorWorkerPool) AddTask(task args) {
	w.queuedTaskC <- task
}

func newWorkerPool(num int) *SensorWorkerPool {
	output := &SensorWorkerPool{}
	output.maxWorker = MaxWorkersAmount
	output.totalSensorsAmount = num
	output.queuedTaskC = make(chan args, num)
	return output
}

func createConn() *grpc_db.SensorStreamClient {
	roundrobinConn, err := grpc.Dial(
		"DUMMY",
		grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [{"round_robin":{}}]}`), // set the initial balancing policy.
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("did not connect: %v\n", err)
	}
	output := grpc_db.NewSensorStreamClient(roundrobinConn)
	return &output
}

func RunSensorWorkingPool(ctx context.Context, max int) {
	//start workerpool
	wp := newWorkerPool(max)
	wp.Run()

	for i := 0; i < max; i++ {
		wp.AddTask(args{createConn()})
	}

	//on another goroutine waiting to the signal handler
	go func(ctx context.Context) {
		<-ctx.Done()
		log.Println("closing sensor worker pool")
		close(wp.queuedTaskC)
	}(ctx)

}
