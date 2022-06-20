package main

import (
	"SensorServer/internal/server"
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/pkg/profile"
	"golang.org/x/sync/errgroup"
)

var (
	verbose    = flag.Bool("v", false, "Verbose mode")
	port       = flag.Int("port", 50051, "GRPC port")
	grpcServer server.ProtocolServer
	sigHandler server.SignalHandler
)

func main() {
	defer profile.Start(profile.TraceProfile, profile.ProfilePath("."), profile.NoShutdownHook).Stop()
	flag.Parse()
	ctx, cancel := context.WithCancel(context.Background())
	g, gctx := errgroup.WithContext(ctx)

	//GRPC server
	g.Go(func() error {
		grpcServer = server.NewServer(verbose)
		grpcServer.RunServer(gctx, *port)
		return nil
	})

	// signal handler
	g.Go(func() error {
		sigHandler = server.NewHandler(cancel, gctx)
		return sigHandler.RunService()
	})

	//wait for all errgroup goroutines
	err := g.Wait()
	if err != nil {
		if errors.Is(err, context.Canceled) {
			fmt.Print("context was canceled\n")
		} else {
			fmt.Printf("received error: %v", err)
		}
	} else {
		fmt.Println("finished shutdown")
	}
	fmt.Println("exit..")
}

//TODO
/*
	1)
	2)
*/
