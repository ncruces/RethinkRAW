package main

import (
	"net/http"
)

func thumbHandler(w http.ResponseWriter, r *http.Request) HTTPResult {
	path := r.URL.Path

	if out, err := previewJPEG(path); err != nil {
		return handleError(err)
	} else {
		w.Header().Add("Content-Type", "image/jpeg")
		w.Write(out)
		return HTTPResult{}
	}
}
