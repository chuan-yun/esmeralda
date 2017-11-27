package controller

import (
	"net/http"
)

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`{"error": "404 Not Found"}`))
}
