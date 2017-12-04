package collector

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"chuanyun.io/esmeralda/collector/storage"
	"chuanyun.io/esmeralda/collector/trace"
	"chuanyun.io/esmeralda/setting"
	"chuanyun.io/esmeralda/util"
	"chuanyun.io/quasimodo/config"
	"github.com/Shopify/sarama"
	"github.com/julienschmidt/httprouter"
	gocache "github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	"github.com/wvanbergen/kafka/consumergroup"
	"golang.org/x/sync/errgroup"
	elastic "gopkg.in/olivere/elastic.v5"
)

type CollectorService struct {
	Cache               *gocache.Cache
	SpansProcessingChan chan *[]trace.Span
	DocumentQueueChan   chan *[]trace.Document
	DocumentQueue       []trace.Document
	Mux                 *sync.Mutex
}

var Service = NewCollectorService()

func NewCollectorService() *CollectorService {

	return &CollectorService{
		Cache:               gocache.New(60*time.Second, 60*time.Second),
		SpansProcessingChan: make(chan *[]trace.Span),
		DocumentQueueChan:   make(chan *[]trace.Document),
		DocumentQueue:       []trace.Document{},
		Mux:                 &sync.Mutex{},
	}
}

func (service *CollectorService) Run(ctx context.Context) error {

	// logCollectMetric := prometheus.NewGaugeVec(
	// 	prometheus.GaugeOpts{
	// 		Name: "kafka_broker_consumer_group_current_offset",
	// 		Help: "Consuming offset of each consumer group/topic/partition based on committed offset",
	// 	},
	// 	[]string{"topic", "partition", "group"},
	// )

	// prometheus.MustRegister(logCollectMetric)

	logrus.Info("Initializing CollectorService")

	group, _ := errgroup.WithContext(ctx)
	group.Go(func() error { return service.queueRoutine(ctx) })
	group.Go(func() error { return service.documentRoutine(ctx) })

	err := group.Wait()

	logrus.Info("Done CollectorService")

	return err
}

func (service *CollectorService) kafkaRoutine(ctx context.Context) error {

	consumerConfig := consumergroup.NewConfig()
	consumerConfig.Offsets.ProcessingTimeout = 5 * time.Second
	if setting.Settings.Kafka.Consumer.Buffer < 0 || setting.Settings.Kafka.Consumer.Buffer > 1024 {
		setting.Settings.Kafka.Consumer.Buffer = 10
	}
	consumerConfig.ChannelBufferSize = setting.Settings.Kafka.Consumer.Buffer
	if config.Config.Kafka.IsResetOffsets {
		consumerConfig.Offsets.ResetOffsets = true
		consumerConfig.Offsets.Initial = sarama.OffsetNewest
	}

	if setting.Settings.Kafka.Zookeeper.Root != "" && setting.Settings.Kafka.Zookeeper.Root != "/" {
		consumerConfig.Zookeeper.Chroot = setting.Settings.Kafka.Zookeeper.Root
	}

	zookeepers := strings.Split(config.Config.Kafka.Zookeeper.Hosts, ",")
	topics := strings.Split(config.Config.Kafka.Topics, ",")
	consumer, consumerError := consumergroup.JoinConsumerGroup(
		setting.Settings.Kafka.Consumer.Group,
		setting.Settings.Kafka.Topics,
		setting.Settings.Kafka.Zookeeper.Servers,
		consumerConfig)

	if consumerError != nil {
		logrus.Fatal(consumerError)
	}

	closeConsumer := func() {
		logrus.Info("main: closing consumer")
		if error := consumer.Close(); error != nil {
			logrus.Fatal(error)
		}
	}
}

func (service *CollectorService) queueRoutine(ctx context.Context) error {
	logrus.Info("Start CollectorService queue routine")

	var assignSpansToQueue = func(spans *[]trace.Span) {

		for _, span := range *spans {
			doc, err := span.AssembleDocument()
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err,
					"span":  span,
				}).Warn(util.Message("span encode to json error"))
				continue
			}
			service.Mux.Lock()
			if len(service.DocumentQueue) < 2 {
				service.DocumentQueue = append(service.DocumentQueue, *doc)
			} else {
				var queue = make([]trace.Document, len(service.DocumentQueue))
				copy(queue, service.DocumentQueue)
				service.DocumentQueueChan <- &queue
				service.DocumentQueue = []trace.Document{}
			}
			service.Mux.Unlock()
		}
	}

	for {
		select {
		case spans := <-Service.SpansProcessingChan:
			assignSpansToQueue(spans)
		case <-ctx.Done():
			logrus.Info("Done collector service queue routine")
			return ctx.Err()
		}
	}
}

func (service *CollectorService) documentRoutine(ctx context.Context) error {
	logrus.Info("Start collector service document routine")

	var bulkSaveDocument = func(documents *[]trace.Document) {

		bulkRequest := setting.Settings.Elasticsearch.Client.Bulk()

		for _, document := range *documents {
			cacheKey := document.IndexName + document.TypeName

			if _, found := service.Cache.Get(cacheKey); !found {
				exists, err := setting.Settings.Elasticsearch.Client.IndexExists(document.IndexName).Do(ctx)
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"error": err,
						"index": document.IndexName,
					}).Warn(util.Message("index exists query error"))
					continue
				}
				if !exists {
					createIndex, err := setting.Settings.Elasticsearch.Client.
						CreateIndex(document.IndexName).
						BodyString(storage.Mappings[document.IndexBaseName]).
						Do(ctx)

					if err != nil {
						logrus.WithFields(logrus.Fields{
							"error": err,
							"index": document.IndexName,
						}).Warn(util.Message("index create error"))
						continue
					}
					if !createIndex.Acknowledged {
						logrus.WithFields(logrus.Fields{
							"error": err,
							"index": document.IndexName,
						}).Warn(util.Message("index create not acknowledged"))
						continue
					}
				}
				service.Cache.Set(cacheKey, true, gocache.DefaultExpiration)
			}

			indexRequest := elastic.NewBulkIndexRequest().
				Index(document.IndexName).
				Type(document.TypeName).
				Doc(document.Payload)

			bulkRequest = bulkRequest.Add(indexRequest)
		}

		bulkResponse, err := bulkRequest.Do(ctx)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Warn(util.Message("bulk save documents error"))

			return
		}
		if bulkResponse == nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Warn(util.Message("bulk save documents response error"))

			return
		}

		indexed := bulkResponse.Indexed()

		if len(indexed) > 0 {
			for _, value := range indexed {
				if value.Status != 201 {
					logrus.WithFields(logrus.Fields{
						"status": value.Status,
						"index":  value.Index,
						"error":  value.Error,
					}).Warn(util.Message("bulk save documents value state error"))
				}
			}
		}
	}

	for {
		select {
		case queue := <-Service.DocumentQueueChan:
			bulkSaveDocument(queue)
		case <-ctx.Done():
			logrus.Info("Done collector service document routine")
			return ctx.Err()
		}
	}
}

func HTTPCollector(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	body := util.RequestBodyToString(r.Body)

	logrus.WithFields(logrus.Fields{
		"size": r.ContentLength,
		"addr": util.IP(r),
	}).Info("trace log statistics")

	spans, err := trace.ToSpans(body)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
			"trace": body,
		}).Warn(util.Message("trace log decode to json error"))

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
