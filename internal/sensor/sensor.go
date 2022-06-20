package sensor

import (
	grpc_db "SensorServer/pkg/grpc_db"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"time"
)

const ()

type sensor struct {
	serial       string
	err          string
	conn         *grpc.ClientConn
	streamClient grpc_db.SensorStreamClient
}

func (s *sensor) sendMeasure() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := s.streamClient.SensorMeasure(ctx, &grpc_db.Measure{
		M:      randMeasure(),
		Serial: s.serial,
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *sensor) Run(ctx context.Context) {

	defer func(roundrobinConn *grpc.ClientConn) {
		err := roundrobinConn.Close()
		if err != nil {
			log.Fatalln(s.err, err)
		}
	}(s.conn)
runLoop:
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("[%s] got signal, exit\n", s.serial)
			break runLoop
		default:
			if err := s.sendMeasure(); err != nil {
				log.Println(s.err, err)
			}
		}
		time.Sleep(time.Second)
	}
}

func Init(sensorSerial string) *sensor {
	sensorErr := fmt.Sprintf("[%s]\t", sensorSerial)

	roundrobinConn, err := grpc.Dial(
		//fmt.Sprintf("multi:///%s,multi:///%s,multi:///%s", addrs[0], addrs[1], addrs[2]),
		"DUMMY",
		grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [{"round_robin":{}}]}`), // set the initial balancing policy.
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("did not connect: %v\n", err)
	}

	// Make another ClientConn with round_robin policy.
	return &sensor{
		serial:       sensorSerial,
		conn:         roundrobinConn,
		streamClient: grpc_db.NewSensorStreamClient(roundrobinConn),
		err:          sensorErr,
	}
}
