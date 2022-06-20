package server

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// SignalHandler interface to represent a server
type SignalHandler interface {
	NewHandler(context.CancelFunc, context.Context)
	RunService() error
}

type InterruptHandler struct {
	sigChan     chan os.Signal
	cancelFunc  context.CancelFunc
	errGroupCtx context.Context
}

func (sh *InterruptHandler) NewHandler(c context.CancelFunc, gctx context.Context) {
	sh.sigChan = make(chan os.Signal, 1)
	signal.Notify(sh.sigChan, os.Interrupt, syscall.SIGTERM)
	sh.cancelFunc = c
	sh.errGroupCtx = gctx
}

func (sh *InterruptHandler) RunService() error {
	select {
	case sig := <-sh.sigChan:
		close(sh.sigChan)
		fmt.Printf("\nReceived signal: %s\n", sig)
		sh.cancelFunc()
		break
	case <-sh.errGroupCtx.Done():
		fmt.Printf("closing signal goroutine\n")
		return sh.errGroupCtx.Err()
	}
	return nil
}

func NewHandler(c context.CancelFunc, gctx context.Context) *InterruptHandler {
	output := &InterruptHandler{}
	output.NewHandler(c, gctx)
	return output
}
