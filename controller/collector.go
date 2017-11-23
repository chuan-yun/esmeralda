package controller

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func Collect(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Write([]byte(`
		<html>
			<head><title>Metrics Exporter</title></head>
			<body>
				<h1>Metrics Exporter</h1>
				<p><a href="./metrics">Metrics</a></p>
			</body>
		</html>`))
}
