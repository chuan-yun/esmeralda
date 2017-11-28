package controller

import (
	"bytes"
	"net/http"

	"chuanyun.io/esmeralda/collector/trace"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
)

func Collect(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	body := buf.String()

	spans, spanJSONError := trace.ToSpans(body)
	if spanJSONError != nil {
		logrus.WithFields(logrus.Fields{
			"error": spanJSONError,
			"trace": body,
		}).Warn("main: trace log decode to json error")
	}
	logrus.Info(spans)
}
