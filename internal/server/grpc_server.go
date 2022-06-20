package server

import (
	"SensorServer/pkg/databases/redis_db"
	grpc_db "SensorServer/pkg/grpc_db"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"net"
)

// ProtocolServer interface to represent a server
type ProtocolServer interface {
	newServer()
	RunServer(context.Context, int)
	cleanup()
}

//interface to represent DB functionalities
type sensorDB interface {
	AddMeasure(string, int32)
	GetInfo(string, int32) string
	DayCleanup()
}

// GrpcServer struct to hold the GRPC's handlers
type GrpcServer struct {
	grpc_db.UnimplementedClientInfoServer   //handle client request
	grpc_db.UnimplementedSensorStreamServer //hande sensors measures
}

var (
	adminIsConnected = false
	sensorCount      = make(chan int64, 1)
	gs               *grpc.Server
	lis              net.Listener
	db               sensorDB
	verbose          *bool
	port             int
)

const (
	adminName = "yochbad"
	adminPass = "123"
)

func returnError(s string) error {
	err := status.Error(codes.Unimplemented, s)
	if err != nil {
		return err
	}
	debug("returnError", s)
	return nil
}

func debug(f string, s string) {
	if *verbose {
		log.Printf("[%s]: %v", f, s)
	}
}

// ClientInfo implementation

func (s *GrpcServer) ConnectClient(ctx context.Context, in *grpc_db.ConnReq) (*grpc_db.ConnRes, error) {
	f := "ConnectClient"
	debug(f, fmt.Sprintf("%v", in))

	if adminIsConnected { //can't connect twice
		debug(f, "adminIsConnected is true")
		return &grpc_db.ConnRes{Res: ""}, returnError("yochbad is already connected")
	}
	if in.UserName != adminName || in.Password != adminPass {
		debug(f, fmt.Sprintf("Wrong credentials:\tin.UserName:%v, in.Password:%v", in.UserName, in.Password))
		return &grpc_db.ConnRes{Res: ""}, returnError("Wrong credentials")
	}
	debug(f, "Connect Success!")
	adminIsConnected = true
	return &grpc_db.ConnRes{Res: "Connected successfully"}, nil
}

func (s *GrpcServer) DisconnectClient(ctx context.Context, in *grpc_db.DisConnReq) (*grpc_db.ConnRes, error) {
	f := "DisconnectClient"
	debug(f, fmt.Sprintf("%v", "enter"))

	if !adminIsConnected { //can't disconnect is not connected first
		debug(f, "adminIsConnected is false, DisconnectClient error")
		return &grpc_db.ConnRes{Res: ""}, returnError("yochbad is not connected")
	}

	debug(f, "Disconnected successfully")
	adminIsConnected = false
	return &grpc_db.ConnRes{Res: "Disconnected successfully"}, nil
}

func (s *GrpcServer) GetInfo(ctx context.Context, in *grpc_db.InfoReq) (*grpc_db.InfoRes, error) {
	f := "GetInfo"
	debug(f, fmt.Sprintf("args:%v", in))
	//unpack request for sensorDB interface
	res := db.GetInfo(in.GetSensorName(), in.GetDayBefore())

	return &grpc_db.InfoRes{Responce: res}, nil
}

// SensorStream implementation

func (s *GrpcServer) ConnectSensor(ctx context.Context, in *grpc_db.ConnSensorReq) (*grpc_db.ConnSensorRes, error) {
	f := "ConnectSensor"
	var num int64
	debug(f, fmt.Sprintf("args:%v", in))
	//get the next serial number and increase by 1 the value to the next
	num = <-sensorCount
	sensorCount <- num + 1
	return &grpc_db.ConnSensorRes{Serial: fmt.Sprintf("sensor_%d", num)}, nil
}

func (s *GrpcServer) SensorMeasure(ctx context.Context, in *grpc_db.Measure) (*grpc_db.MeasureRes, error) {
	f := "SensorMeasure"
	debug(f, fmt.Sprintf("%d\tgot measure=%d from %s", port, in.GetM(), in.GetSerial()))
	db.DayCleanup()
	//unpack request for sensorDB interface
	db.AddMeasure(in.GetSerial(), in.GetM())
	return &grpc_db.MeasureRes{}, nil
}

func (s *GrpcServer) newServer() {
	gs = grpc.NewServer()
}

// RunServer implementation of ProtocolServer interface
func (s *GrpcServer) RunServer(parentCtx context.Context, serverPort int) {
	port = serverPort
	//attach the goroutine's context to the goroutine of the server
	_, cancelServer := context.WithCancel(parentCtx)
	defer cancelServer()

	var err error
	lis, err = net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	//gs = grpc.NewServer()
	grpc_db.RegisterSensorStreamServer(gs, &GrpcServer{})
	grpc_db.RegisterClientInfoServer(gs, &GrpcServer{})
	sensorCount <- 1

	//DB - used sensorDB interface
	//db = In_memo_db.SensorMap()
	db = redis_db.New()

	if db == nil {
		log.Fatalf("DB was not installed correctly")
	}

	log.Printf("server listening at %v", lis.Addr())

	select {
	case <-parentCtx.Done():
		log.Println("shutting GRPC server down")
		s.cleanup()
	default:
		go func() {
			if err := gs.Serve(lis); err != nil {
				log.Fatalf("%v\n", err)
			}
		}()
	}
}

func (s *GrpcServer) cleanup() {
	adminIsConnected = false
	gs.GracefulStop()
	close(sensorCount)
}

func NewServer(v *bool) *GrpcServer {
	verbose = v
	gs := &GrpcServer{}
	gs.newServer()
	return gs
}
