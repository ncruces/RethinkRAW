package main

import (
	"net/http"
)

func thumbHandler(w http.ResponseWriter, r *http.Request) HttpResult {
	path := r.URL.Path

	if out, err := getThumb(path); err != nil {
		return handleError(err)
	} else {
		w.Header().Add("Content-Type", "image/jpeg")
		w.Write(out)
		return HttpResult{}
	}
}
