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

var SpansProcessingChan = make(chan *[]trace.Span)

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

	select {
	case SpansProcessingChan <- spans:
		w.Write([]byte(`{"error": "SpansProcessingChan <- spans"}`))
	default:
		w.Write([]byte(`{"error": "default"}`))
	}
}

func SpansToDoc() {
	go func() {
		for spans := range SpansProcessingChan {

			for index := range *spans {
				doc, asError := (*spans)[index].AssembleDocument()
				if asError != nil {
					logrus.WithFields(logrus.Fields{
						"error": asError,
						"span":  (*spans)[index],
					}).Warn(util.Message("span encode to json error"))
					continue
				}
				logrus.Info(doc)
			}
		}
	}()
}

// func saveDoc(docCh <-chan elasticsearch.Document) error {
// 	go func() {
// 		for document := range SpansProcessingChan {
// 			cacheKey := document.IndexName + document.TypeName

// 			_, found := indexCache.Get(cacheKey)
// 			if found {
// 				// logrus.Info("main: index:" + indexName + " exists.")
// 			} else {
// 				// Use the IndexExists service to check if a specified index exists.
// 				exists, err := elasticsearchClient.IndexExists(document.IndexName).Do(ctx)
// 				if err != nil {
// 					logrus.Fatal(err)
// 				}
// 				if !exists {

// 					createIndex, err := elasticsearchClient.CreateIndex(document.IndexName).BodyString(elasticsearch.Mappings[document.IndexBaseName]).Do(ctx)
// 					if err != nil {
// 						logrus.Warn(err)
// 						continue
// 					}
// 					if !createIndex.Acknowledged {
// 						// Not acknowledged
// 					}
// 				}
// 				indexCache.Set(cacheKey, true, cache.DefaultExpiration)

// 				// aliasService := elastic.NewAliasService(elasticsearchClient)
// 				// aliasService.Add(document.IndexName, "alias-"+document.IndexName)
// 			}

// 			indexRequest := elastic.NewBulkIndexRequest().Index(document.IndexName).Type(document.TypeName).Doc(document.Payload)
// 			bulkRequest = bulkRequest.Add(indexRequest)
// 		}

// 		bulkResponse, err := bulkRequest.Do(ctx)
// 		if err != nil {
// 			logrus.Fatal(err)
// 		}
// 		if bulkResponse == nil {
// 			logrus.Fatal("main: expected bulkResponse to be != nil; got nil")
// 		}

// 		indexed := bulkResponse.Indexed()

// 		if len(indexed) > 0 {
// 			for _, value := range indexed {
// 				if value.Status != 201 {
// 					logrus.Error("main: document bulk index error:" + value.Index)
// 				}
// 			}
// 		}
// 	}()

// 	return nil
// }
