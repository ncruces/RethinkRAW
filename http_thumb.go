package main

import (
	"net/http"
)

func thumbHandler(w http.ResponseWriter, r *http.Request) HTTPResult {
	path := fromURLPath(r.URL.Path)

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "max-age=60")
	if r := cacheHeaders(path, r.Header, w.Header()); r.Done() {
		return r
	}

	if r.Method == "HEAD" {
		return HTTPResult{}
	}

	if out, err := previewJPEG(path); err != nil {
		return HTTPResult{Error: err}
	} else {
		w.Write(out)
		return HTTPResult{}
	}
}
