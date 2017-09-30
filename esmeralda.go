package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"time"

	"chuanyun.io/quasimodo/scheduler"

	"chuanyun.io/quasimodo/analysis"
	"chuanyun.io/quasimodo/config"
	"chuanyun.io/quasimodo/elasticsearch"
	"chuanyun.io/quasimodo/pack"
	"chuanyun.io/quasimodo/trace"

	"github.com/Shopify/sarama"
	cache "github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/wvanbergen/kafka/consumergroup"
	elastic "gopkg.in/olivere/elastic.v5"

	"net/http"
	_ "net/http/pprof"

	"github.com/sirupsen/logrus"

	_ "github.com/garyburd/redigo/redis"
	_ "github.com/spf13/cobra"
)

var (
	version = flag.Bool("version", false, "output version information and exit.")

	GitTag    = "2000.01.01.release"
	BuildTime = "2000-01-01T00:00:00+0800"

	indexCache = cache.New(24*time.Hour, 24*time.Hour)

	elasticsearchClient *elastic.Client
	ctx                 context.Context
	cancel              context.CancelFunc
)

func init() {

	flag.StringVar(&config.Config.Kafka.GroupID, "kafka.group.id", "", "kafka consumer group id")
	flag.StringVar(&config.Config.Kafka.Topics, "kafka.topics", "", "comma-separated kafka topics")
	flag.IntVar(&config.Config.Kafka.BufferSize, "kafka.buffer", 10, "kafka consumer buffer size")
	flag.StringVar(&config.Config.Kafka.Zookeeper.Hosts, "zookeeper.addr", "", "a comma-separated zookeeper connection string (e.g. `zookeeper1.local:2181,zookeeper2.local:2181`)")
	flag.StringVar(&config.Config.Kafka.Zookeeper.Path, "zookeeper.path", "/", "kafka broker's root path in zookeeper")
	flag.StringVar(&config.Config.Elasticsearch.Hosts, "elasticsearch.hosts", "", "elasticsearch's hosts(e.g. `http://10.209.26.171:11520,http://10.209.26.172:11520`)")
	flag.StringVar(&config.Config.Elasticsearch.Username, "elasticsearch.username", "", "elasticsearch's username(HTTP Basic Auth)")
	flag.StringVar(&config.Config.Elasticsearch.Password, "elasticsearch.password", "", "elasticsearch's password(HTTP Basic Auth)")
	flag.StringVar(&config.Config.Prometheus.Port, "exporter.port", "10301", "the address to listen on for prometheus monitor")
	flag.StringVar(&config.Config.Log.Level, "log.level", "Info", "logging levels: Debug, Info, Warning, Error, Fatal, Panic")
	flag.StringVar(&config.Config.Application.Env, "app.env", "production", "application env: development|testing|staging|production")
	flag.StringVar(&config.Config.MySQL.DSN, "mysql.dsn", "chuanyun:TianShang1Ge**@tcp(10.213.58.181:13306)/chuanyun?charset=utf8mb4", "mysql data source name")
	flag.IntVar(&config.Config.Elasticsearch.BulkSize, "elasticsearch.bulk.size", 2000, "elasticsearch bulk document size(0 <= size <= 5000)")
	flag.StringVar(&config.Config.Gateway.TranslatorURL, "gateway.url", "", "url transfer gateway address")
	flag.BoolVar(&config.Config.Module.DivideEnable, "module.enable", true, "enable module divide feature")
	flag.IntVar(&config.Config.Module.Threshold, "module.threshold", 5, "module divide threshold")
	flag.BoolVar(&config.Config.Kafka.IsResetOffsets, "kafka.resetOffsets", false, "resets the offsets for the consumergroup so that it won't resume from where it left off previously")

	// profile := flag.Bool("profile", false, "Turn on pprof profiling")
	// profilePort := flag.Int("profile-port", 6060, "Define custom port for profiling")

	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	level, levelError := logrus.ParseLevel(config.Config.Log.Level)
	if levelError != nil {
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)

	sarama.Logger = logrus.StandardLogger()

	ctx, cancel = context.WithCancel(context.Background())
}

func exporter() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`
		<html>
			<head><title>Chuanyun Quasimodo Exporter</title></head>
			<body>
				<h1>Chuanyun Quasimodo Exporter</h1>
				<p><a href="/metrics">Metrics</a></p>
			</body>
		</html>`))
	})
	http.Handle("/metrics", promhttp.Handler())
	logrus.Fatal(http.ListenAndServe(":"+config.Config.Prometheus.Port, nil))
}

func interrupt() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Kill, os.Interrupt)
	<-signals
	logrus.Info("main: initiating shutdown of quasimodo...")
	cancel()
}

func versionInfo() {
	fmt.Println("Quasimodo")
	fmt.Println("    version: " + GitTag + ", build: " + BuildTime)
	fmt.Println("    Copyright (C) 2017 chuanyun.io.")
}

func main() {

	defer func() {
		cancel()
	}()

	runtime.GOMAXPROCS(runtime.NumCPU())

	if len(os.Args) <= 1 {
		flag.Usage()
		return
	}
	flag.Parse()
	if *version {
		versionInfo()
		os.Exit(0)
	}
	if config.Config.Kafka.GroupID == "" || config.Config.Kafka.Topics == "" || config.Config.Kafka.Zookeeper.Hosts == "" || config.Config.Kafka.Zookeeper.Path == "" || config.Config.Elasticsearch.Hosts == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	elasticsearchHosts := strings.Split(config.Config.Elasticsearch.Hosts, ",")
	if len(elasticsearchHosts) <= 0 {
		logrus.Panic("there is no elasticsearch host valid")
	}

	var elasticsearchOptions []elastic.ClientOptionFunc
	elasticsearchOptions = append(elasticsearchOptions, elastic.SetURL(elasticsearchHosts...))
	if config.Config.Elasticsearch.Username != "" && config.Config.Elasticsearch.Password != "" {
		elasticsearchOptions = append(elasticsearchOptions, elastic.SetBasicAuth(config.Config.Elasticsearch.Username, config.Config.Elasticsearch.Password))
	}
	var err error
	elasticsearchClient, err = elasticsearch.GetElasticsearchClient()
	if err != nil {
		logrus.Panic(err)
	}

	if config.Config.Elasticsearch.BulkSize < 0 || config.Config.Elasticsearch.BulkSize > 5000 {
		config.Config.Elasticsearch.BulkSize = 2000
	}

	go exporter()
	go interrupt()

	if config.Config.Module.DivideEnable {
		go scheduler.Start()
	}

	consumerConfig := consumergroup.NewConfig()
	consumerConfig.Offsets.ProcessingTimeout = 5 * time.Second
	if config.Config.Kafka.BufferSize < 0 || config.Config.Kafka.BufferSize > 1024 {
		config.Config.Kafka.BufferSize = 10
	}
	consumerConfig.ChannelBufferSize = config.Config.Kafka.BufferSize
	if config.Config.Kafka.IsResetOffsets {
		consumerConfig.Offsets.ResetOffsets = true
		consumerConfig.Offsets.Initial = sarama.OffsetNewest
	}

	if config.Config.Kafka.Zookeeper.Path != "" && config.Config.Kafka.Zookeeper.Path != "/" {
		consumerConfig.Zookeeper.Chroot = config.Config.Kafka.Zookeeper.Path
	}

	zookeepers := strings.Split(config.Config.Kafka.Zookeeper.Hosts, ",")
	topics := strings.Split(config.Config.Kafka.Topics, ",")
	consumer, consumerError := consumergroup.JoinConsumerGroup(
		config.Config.Kafka.GroupID,
		topics,
		zookeepers,
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

	go func() {
		for err := range consumer.Errors() {
			logrus.Info(err)
		}
	}()

	go func(consumer *consumergroup.ConsumerGroup) {
		<-ctx.Done()
		closeConsumer()
	}(consumer)

	kafkaOffsetMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kafka_broker_consumer_group_current_offset",
			Help: "Consuming offset of each consumer group/topic/partition based on committed offset",
		},
		[]string{"topic", "partition", "group"},
	)
	prometheus.MustRegister(kafkaOffsetMetric)

	offsets := make(map[string]map[int32]int64)

	var consumed uint64

	// messageCh := make(chan *sarama.ConsumerMessage)

	// go func() {
	// 	for message := range consumer.Messages() {

	// 		messageCh <- message

	// 		consumer.CommitUpto(message)
	// 		consumed++

	// 		if offsets[message.Topic] == nil {
	// 			offsets[message.Topic] = make(map[int32]int64)
	// 		}
	// 		if offsets[message.Topic][message.Partition] != 0 && offsets[message.Topic][message.Partition] != message.Offset-1 {
	// 			logrus.WithFields(logrus.Fields{
	// 				"topic":           message.Topic,
	// 				"partition":       message.Partition,
	// 				"expected_offset": offsets[message.Topic][message.Partition] + 1,
	// 				"found_offset":    message.Offset,
	// 				"diff_offset":     message.Offset - offsets[message.Topic][message.Partition] + 1,
	// 			}).Info("main: consumed message offset")
	// 		}

	// 		labels := []string{message.Topic, fmt.Sprintf("%d", message.Partition), config.Config.Kafka.GroupID}

	// 		kafkaOffsetMetric.WithLabelValues(labels...).Set(float64(message.Offset))

	// 		offsets[message.Topic][message.Partition] = message.Offset

	// 		if consumed%1000 == 0 {
	// 			logrus.WithFields(logrus.Fields{
	// 				"offsets": offsets,
	// 			}).Info("main: consumed message offset")
	// 		}
	// 	}

	// }()

	// saveDoc(spansToDoc(messageToSpans(messageCh)))
}

func messageToSpans(messages <-chan *sarama.ConsumerMessage) <-chan *[]trace.Span {
	out := make(chan *[]trace.Span)
	go func() {
		for message := range messages {

			traceLog, err := pack.GetMessageBody(message.Value)
			if err != nil && traceLog == "" {
				traceLog = string(message.Value[:])
			}

			spans, spanJSONError := trace.ToSpans(traceLog)
			if spanJSONError != nil {
				logrus.WithFields(logrus.Fields{
					"error": spanJSONError,
					"trace": string(message.Value[:]),
				}).Warn("main: trace log decode to json error")
				continue
			}

			adjustSpansError := trace.AdjustSpans(spans)
			if adjustSpansError != nil {
				logrus.WithFields(logrus.Fields{
					"error": adjustSpansError,
					"spans": spans,
				}).Warn("main: spans adjust error")
				continue
			}
			out <- spans
		}
		close(out)
	}()

	return out
}

func spansToDoc(spansCh <-chan *[]trace.Span) <-chan elasticsearch.Document {
	out := make(chan elasticsearch.Document)
	go func() {
		for spans := range spansCh {

			for index := range *spans {
				doc, asError := (*spans)[index].AssembleDocument()
				if asError != nil {
					logrus.WithFields(logrus.Fields{
						"error": asError,
						"span":  (*spans)[index],
					}).Warn("main: span encode to json error")
					continue
				}

				out <- doc
			}

			docs, analyseSpansError := analysis.AnalyseSpans(spans)
			if analyseSpansError != nil {
				logrus.WithFields(logrus.Fields{
					"error": analyseSpansError,
					"spans": spans,
				}).Warn("main: spans analysis error")
				continue
			}
			if len(docs) > 0 {
				for _, doc := range docs {
					out <- doc
				}
			}
		}
		close(out)
	}()

	return out
}

func saveDoc(docCh <-chan elasticsearch.Document) error {
	go func() {
		bulkRequest := elasticsearchClient.Bulk()
		for document := range docCh {
			cacheKey := document.IndexName + document.TypeName

			_, found := indexCache.Get(cacheKey)
			if found {
				// logrus.Info("main: index:" + indexName + " exists.")
			} else {
				// Use the IndexExists service to check if a specified index exists.
				exists, err := elasticsearchClient.IndexExists(document.IndexName).Do(ctx)
				if err != nil {
					logrus.Fatal(err)
				}
				if !exists {

					createIndex, err := elasticsearchClient.CreateIndex(document.IndexName).BodyString(elasticsearch.Mappings[document.IndexBaseName]).Do(ctx)
					if err != nil {
						logrus.Warn(err)
						continue
					}
					if !createIndex.Acknowledged {
						// Not acknowledged
					}
				}
				indexCache.Set(cacheKey, true, cache.DefaultExpiration)

				// aliasService := elastic.NewAliasService(elasticsearchClient)
				// aliasService.Add(document.IndexName, "alias-"+document.IndexName)
			}

			indexRequest := elastic.NewBulkIndexRequest().Index(document.IndexName).Type(document.TypeName).Doc(document.Payload)
			bulkRequest = bulkRequest.Add(indexRequest)
		}

		bulkResponse, err := bulkRequest.Do(ctx)
		if err != nil {
			logrus.Fatal(err)
		}
		if bulkResponse == nil {
			logrus.Fatal("main: expected bulkResponse to be != nil; got nil")
		}

		indexed := bulkResponse.Indexed()

		if len(indexed) > 0 {
			for _, value := range indexed {
				if value.Status != 201 {
					logrus.Error("main: document bulk index error:" + value.Index)
				}
			}
		}
	}()

	return nil
}

func merge(cs ...<-chan int) <-chan int {
	var wg sync.WaitGroup
	out := make(chan int)

	// Start an output goroutine for each input channel in cs.  output
	// copies values from c to out until c is closed, then calls wg.Done.
	output := func(c <-chan int) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
