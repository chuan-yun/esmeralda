package collector

import (
	"net/http"

	"chuanyun.io/esmeralda/collector/trace"
	"chuanyun.io/esmeralda/util"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
)

var SpansProcessingChan = make(chan *[]trace.Span)

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
	case SpansProcessingChan <- spans:
		w.Write([]byte(`{"msg": "SpansProcessingChan <- spans"}`))
	default:
		w.Write([]byte(`{"msg": "default"}`))
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
