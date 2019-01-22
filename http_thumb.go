package main

import (
	"net/http"
)

func thumbHandler(w http.ResponseWriter, r *http.Request) HTTPResult {
	path := r.URL.Path

	res := cacheHeaders(path, r.Header, w.Header())
	if res.Status == 0 {
		if out, err := previewJPEG(path); err != nil {
			return handleError(err)
		} else {
			w.Header().Set("Cache-Control", "max-age=60")
			w.Header().Set("Content-Type", "image/jpeg")
			w.Write(out)
		}
	}
	return res
}
