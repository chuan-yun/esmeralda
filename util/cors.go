package util

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	// "github.com/cristiangraz/kumi"
	// "fmt"
)

func CORS(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		// if len(ps) > 0 {
		// 	p := make(map[string]string, len(ps))
		// 	for _, v := range ps {
		// 		p[v.Key] = v.Value
		// 	}
		// 	r = kumi.SetParams(r, p)
		// }
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")
		h(w, r, ps)
	}
}
