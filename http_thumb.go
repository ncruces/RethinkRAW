package main

import (
	"net/http"
)

func thumbHandler(w http.ResponseWriter, r *http.Request) HTTPResult {
	path := r.URL.Path

	if r := cacheHeaders(path, r.Header, w.Header()); r.Done() {
		return r
	}

	if out, err := previewJPEG(path); err != nil {
		return HTTPResult{Error: err}
	} else {
		w.Header().Set("Cache-Control", "max-age=60")
		w.Header().Set("Content-Type", "image/jpeg")
		w.Write(out)
		return HTTPResult{}
	}
}
