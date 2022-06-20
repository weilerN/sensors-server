package main

import (
	"SensorServer/internal/sensor"
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	numOfSensors = flag.Int("n", 10, "number of sensors in simulator")
)

func main() {
	flag.Parse()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg := &sync.WaitGroup{}

	go func() {
		wg.Add(1)
		defer wg.Done()
		sensor.RunSensorWorkingPool(ctx, *numOfSensors)
	}()

	//signal handler
	go func() {
		wg.Add(1)
		defer wg.Done()
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		sig := <-sigChan
		log.Println("\ninterrupted by:", sig)
		cancel()
	}()

	log.Println("All sensors are up")
	wg.Wait()
	log.Println("finished simulation")

}
