package main

import (
	"fmt"
	"github.com/bojand/ghz/runner"
	"os"
	"sync"
)

func runStream() int {
	report, err := runner.Run(
		"SensorServer.SensorStream.SensorMeasure",
		"localhost:50051",
		runner.WithProtoFile("./pkg/grpc_db/grpc_db.proto", []string{}),
		runner.WithDataFromFile("./internal/sensor/big_data.json"),
		runner.WithInsecure(true),
		runner.WithTotalRequests(30000),
	)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	return int(report.Count)
}
func main() {

	wg := &sync.WaitGroup{}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(num int) {
			fmt.Printf("Streamer %d start\n", num)
			defer wg.Done()
			fmt.Printf("Streamer %d finished, total:%d\n", num, runStream())
		}(i)
	}

	wg.Wait()
	fmt.Println("finished streaming")

	//p := printer.ReportPrinter{
	//	Out:    os.Stdout,
	//	Report: report,
	//}

	//if err2 := p.Print("pretty"); err2 != nil {
	//	fmt.Println(err2)
	//}

	//fmt.Println(report)
}
