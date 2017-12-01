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
	context             context.Context
	Cache               *gocache.Cache
	SpansProcessingChan chan *[]trace.Span
	DocumentQueueChan   chan []trace.Document
	DocumentQueue       DocumentQueue
}

type DocumentQueue struct {
	Queue []trace.Document
	Mux   *sync.Mutex
}

var Service = NewCollectorService()

func NewCollectorService() *CollectorService {
	return &CollectorService{
		Cache:               gocache.New(60*time.Second, 60*time.Second),
		SpansProcessingChan: make(chan *[]trace.Span),
		DocumentQueueChan:   make(chan []trace.Document),
		DocumentQueue: DocumentQueue{
			Queue: []trace.Document{},
			Mux:   &sync.Mutex{},
		},
	}
}

func (service *CollectorService) Run(ctx context.Context) error {

	service.context = ctx
	logrus.Info("Initializing CollectorService")

	group, _ := errgroup.WithContext(ctx)
	group.Go(func() error { return service.queueRoutine(ctx) })
	group.Go(func() error { return service.documentRoutine(ctx) })

	err := group.Wait()

	logrus.Info("Done CollectorService")

	return err
}

func (service *CollectorService) queueRoutine(ctx context.Context) error {
	for {
		select {
		case spans := <-Service.SpansProcessingChan:
			for index := range *spans {
				doc, err := (*spans)[index].AssembleDocument()
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"error": err,
						"span":  (*spans)[index],
					}).Warn(util.Message("span encode to json error"))
					continue
				}
				service.DocumentQueue.Mux.Lock()
				if len(service.DocumentQueue.Queue) < 2 {
					service.DocumentQueue.Queue = append(service.DocumentQueue.Queue, *doc)
				} else {
					var queue = make([]trace.Document, len(service.DocumentQueue.Queue))
					copy(queue, service.DocumentQueue.Queue)
					service.DocumentQueueChan <- queue
					service.DocumentQueue.Queue = []trace.Document{}

				}
				service.DocumentQueue.Mux.Unlock()
			}
		case <-ctx.Done():
			logrus.Info(util.Message("Done SpansToDocumentQueue"))
			return ctx.Err()
		}
	}
}

func (service *CollectorService) documentRoutine(ctx context.Context) error {
	logrus.Info(util.Message("start"))
	for {
		select {
		case queue := <-Service.DocumentQueueChan:
			logrus.WithFields(logrus.Fields{
				"queue": queue,
			}).Info(queue)
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
	case Service.SpansProcessingChan <- spans:
		w.Write([]byte(`{"msg": "SpansProcessingChan <- spans"}`))
	default:
		w.Write([]byte(`{"msg": "default"}`))
	}
}
