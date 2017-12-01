package collector

import (
	"context"
	"net/http"
	"sync"
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
	DocumentQueueChan   chan []trace.Document
	Cache               *gocache.Cache
}

var Collector = newCollectorService()

var DocumentQueue struct {
	Queue []trace.Document
	Mux   sync.Mutex
}

func newCollectorService() *CollectorService {
	collectorService := &CollectorService{}
	collectorService.Cache = gocache.New(60*time.Second, 60*time.Second)
	collectorService.SpansProcessingChan = make(chan *[]trace.Span)

	return collectorService
}

func RunCollectorService(ctx context.Context) error {
	logrus.Info("Initializing CollectorService")

	group, _ := errgroup.WithContext(ctx)
	group.Go(func() error { return BulkSaveDocumentRoutine(ctx) })
	group.Go(func() error { return spansToDocumentQueueRoutine(ctx) })

	err := group.Wait()

	logrus.Info("Done CollectorService")

	return err
}

func spansToDocumentQueueRoutine(ctx context.Context) error {
	for {
		select {
		case spans := <-Collector.SpansProcessingChan:
			logrus.Info(util.Message(""))
			for index := range *spans {
				logrus.Info(util.Message(""))
				doc, err := (*spans)[index].AssembleDocument()
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"error": err,
						"span":  (*spans)[index],
					}).Warn(util.Message("span encode to json error"))
					continue
				}
				logrus.Info(util.Message(""))
				DocumentQueue.Mux.Lock()
				if len(DocumentQueue.Queue) < 2 {
					logrus.Info(util.Message(""))
					DocumentQueue.Queue = append(DocumentQueue.Queue, *doc)
				} else {
					logrus.Info(util.Message(""))
					var c = []trace.Document{}
					copy(c, DocumentQueue.Queue)
					Collector.DocumentQueueChan <- c
					DocumentQueue.Queue = []trace.Document{}
					logrus.Info(util.Message(""))
				}
				logrus.Info(util.Message(""))
				DocumentQueue.Mux.Unlock()
			}
		case <-ctx.Done():
			logrus.Info(util.Message("Done SpansToDocumentQueue"))
			return ctx.Err()
		}
	}
}

func BulkSaveDocumentRoutine(ctx context.Context) error {
	logrus.Info(util.Message("start"))
	for {
		select {
		case queue := <-Collector.DocumentQueueChan:
			logrus.Info(util.Message(""))
			logrus.Info(queue)
		case <-ctx.Done():
			logrus.Info(util.Message("Done BulkSaveDocument"))
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
