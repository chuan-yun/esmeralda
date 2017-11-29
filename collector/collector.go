package collector

import (
	"context"
	"net/http"

	"chuanyun.io/esmeralda/collector/trace"
	"chuanyun.io/esmeralda/util"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

var SpansChan = make(chan *[]trace.Span)

type CollectorServerImpl struct {
	context       context.Context
	shutdownFn    context.CancelFunc
	childRoutines *errgroup.Group
}

func (me *CollectorServerImpl) Start() {

}

func (me *CollectorServerImpl) Shutdown(code int, reason string) {

}

// func NewCollectorServer() server.Server {
// 	rootCtx, shutdownFn := context.WithCancel(context.Background())
// 	childRoutines, childCtx := errgroup.WithContext(rootCtx)

// 	return &CollectorServerImpl{
// 		context:       childCtx,
// 		shutdownFn:    shutdownFn,
// 		childRoutines: childRoutines,
// 	}
// }

func HTTPCollector(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	body := util.RequestBodyToString(r.Body)

	logrus.WithFields(logrus.Fields{
		"size": r.ContentLength,
		"addr": util.IP(r),
	}).Info(util.Message("trace log statistics"))

	spans, err := trace.ToSpans(body)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
			"trace": body,
		}).Warn("main: trace log decode to json error")
	}

	logrus.Info(spans)
	// SpansChan <- spans
}
