package collector

import (
	"context"

	"chuanyun.io/esmeralda/collector/trace"
	"golang.org/x/sync/errgroup"
)

var spansChan = make(chan *[]trace.Span)

type CollectorServerImpl struct {
	context       context.Context
	shutdownFn    context.CancelFunc
	childRoutines *errgroup.Group
}

func (me *CollectorServerImpl) Start() {

}

func (me *CollectorServerImpl) Shutdown(code int, reason string) {

}

func NewCollectorServer() Server {
	rootCtx, shutdownFn := context.WithCancel(context.Background())
	childRoutines, childCtx := errgroup.WithContext(rootCtx)

	return &CollectorServerImpl{
		context:       childCtx,
		shutdownFn:    shutdownFn,
		childRoutines: childRoutines,
	}
}
