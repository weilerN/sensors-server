package client

import (
	grpc_db "SensorServer/pkg/grpc_db"
	"context"
	"errors"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"time"
)

const (
	UserExit              = "user Exit"
	loginConnectedMessage = "Connected successfully"
	loginCredentialsError = "Wrong credentials"
	alreadyConnectedError = "yochbad is already connected"
)

var (
	isConnected      bool
	clientInfoClient grpc_db.ClientInfoClient
	addr             = flag.String("addr", "localhost:50051", "the address to connect to")
	verbose          = flag.Bool("v", false, "Verbose mode")
)

func dayOpt() int {
	var opt int
	for {
		fmt.Println("\nplease enter a number between 1-6 (day before today)\n10 - to exit")
		_, err := fmt.Scanf("%d", &opt)
		MyPanic(err)
		switch {
		case 0 < opt && opt < 7:
			fmt.Println("OPT:", opt)
			return opt
		case opt == 10:
			return -1
		default:
			fmt.Println("Illegal option")
		}
	}
}

/*
	return values:
	1-6 :	number of days before today
	8	:	all week
	9	:	today
*/
func showDMenu() int {
	var opt int
	for {
		fmt.Printf("\nchoose day option\n1)\tshow by day - all past week\n2)\tshow by day - specific day\n3)\tshow by day - today\n5)\texit\n")
		_, err := fmt.Scanf("%d", &opt)
		MyPanic(err)
		switch opt {
		case 1:
			return 8
		case 2:
			return dayOpt()
		case 3:
			return 9
		case 5:
			return -1
		default:
			fmt.Println("Illegal option")
		}
	}
}

func ShowMainMenu() int {
	var opt int
	for {
		fmt.Printf("\nchoose day option\n1)\tget info\n2)\tdisconnect\n3)\texit\n")
		_, err := fmt.Scanf("%d", &opt)
		MyPanic(err)
		switch opt {
		case 1, 2, 3:
			return opt
		default:
			fmt.Println("Illegal option")
		}
	}
}

func sensorOpt() string {
	var output string
	fmt.Println("\nenter the sensor name (for example: sensor_1)")
	_, err := fmt.Scanf("%s", &output)
	MyPanic(err)
	return output
}

func showSMenu() string {
	var opt int
	for {
		fmt.Printf("\nplease choose an option for sensors:\n1)\tshow by sensor - all sensors\n2)\tshow by sensor - specific sensor\n5)\texit\n")
		_, err := fmt.Scanf("%d", &opt)
		MyPanic(err)
		switch opt {
		case 1:
			return "all"
		case 2:
			return sensorOpt()
		case 5:
			return UserExit
		default:
			fmt.Println("Illegal option")
		}
	}
}

func showMenu() (int32, string) {
	d := showDMenu()
	if d == -1 { //if already want to quit - exit without further menu options
		return int32(d), ""
	}
	s := showSMenu()
	return int32(d), s
}

func createRequest() *grpc_db.InfoReq {
	d, s := showMenu()
	return &grpc_db.InfoReq{DayBefore: d, SensorName: s}
}

func ClientMenu() (*grpc_db.InfoReq, error) {
	cr := createRequest()
	if cr.SensorName == UserExit || cr.DayBefore == -1 {
		return nil, errors.New(UserExit)
	}
	return cr, nil
}

func ConnectClient() *grpc.ClientConn {
	// Set up a connection to the server.
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	clientInfoClient = grpc_db.NewClientInfoClient(conn)
	return conn
}

func disconnectClient(ctx context.Context) {
	r, err := clientInfoClient.DisconnectClient(ctx, &grpc_db.DisConnReq{})
	if err != nil {
		fmt.Println("Error: ", UnpackError(err))
	} else {
		log.Println(r.GetRes())
		isConnected = false
	}
}

func MenuLoop() {
forLoop:
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		if !isConnected {
			cr := NewConnReq()
			r, err := clientInfoClient.ConnectClient(ctx, cr)
			isConnected = VerifyLogin(r, err)
		} else {
			switch ShowMainMenu() {
			case 1: //got info
				ir, err := ClientMenu()
				if err != nil && fmt.Sprintf("%v", err) == UserExit {
					continue //got error if userExit from menu
				}
				res, err := clientInfoClient.GetInfo(ctx, ir)
				if err != nil {
					fmt.Println(err)
				} else { //got response from server
					Debug("GetInfoRes", res.GetResponce())
					PrintResult(res.GetResponce())
				}
			case 2: //disconnect
				disconnectClient(ctx)
			case 3: //exit
				cancel()
				disconnectClient(ctx) //disconnect by default on exit - my design
				break forLoop
			default:
			}
		}
		cancel() //close the current iteration context
	}
}
