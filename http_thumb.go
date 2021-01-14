package main

import (
	"net/http"
)

func thumbHandler(w http.ResponseWriter, r *http.Request) HTTPResult {
	if r := sendAllowed(w, r, "GET", "HEAD"); r.Done() {
		return r
	}
	path := fromURLPath(r.URL.Path)

	w.Header().Set("Cache-Control", "max-age=60")
	if r := sendCached(w, r, path); r.Done() {
		return r
	}

	w.Header().Set("Content-Type", "image/jpeg")
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
