package collector

import (
	"context"
	"net/http"
	"time"

	"chuanyun.io/esmeralda/collector/trace"
	"chuanyun.io/esmeralda/util"
	"github.com/julienschmidt/httprouter"
	gocache "github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type CollectorService struct {
	SpansProcessingChan chan *[]trace.Span
	Cache               *gocache.Cache
}

var Collector = newCollectorService()

func newCollectorService() *CollectorService {
	collectorService := &CollectorService{}
	collectorService.Cache = gocache.New(60*time.Second, 60*time.Second)
	collectorService.SpansProcessingChan = make(chan *[]trace.Span)

	return collectorService
}

func (me *CollectorService) Run(ctx context.Context) error {
	logrus.Info("Initializing CollectorService")

	group, _ := errgroup.WithContext(ctx)
	group.Go(func() error { return me.start(ctx) })

	err := group.Wait()

	logrus.Info("Done CollectorService")

	return err
}

func (service *CollectorService) start(ctx context.Context) error {

	for {
		select {
		case spans := <-Collector.SpansProcessingChan:
			for index := range *spans {
				doc, err := (*spans)[index].AssembleDocument()
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"error": err,
						"span":  (*spans)[index],
					}).Warn(util.Message("span encode to json error"))
					continue
				}
				logrus.Info(doc)
			}
		case <-ctx.Done():
			logrus.Info(util.Message("done"))
			return ctx.Err()
		}
	}
}

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

		w.Write([]byte(`{"msg": "error trace log"}`))

		return
	}

	select {
	case Collector.SpansProcessingChan <- spans:
		w.Write([]byte(`{"msg": "SpansProcessingChan <- spans"}`))
	default:
		w.Write([]byte(`{"msg": "default"}`))
	}
}
