package main

import "net/http"

func thumbHandler(w http.ResponseWriter, r *http.Request) httpResult {
	if r := sendAllowed(w, r, "GET", "HEAD"); r.Done() {
		return r
	}
	prefix := getPathPrefix(r)
	path := fromURLPath(r.URL.Path, prefix)

	w.Header().Set("Cache-Control", "max-age=60")
	if r := sendCached(w, r, path); r.Done() {
		return r
	}

	w.Header().Set("Content-Type", "image/jpeg")
	if r.Method == "HEAD" {
		return httpResult{}
	}

	if out, err := previewJPEG(r.Context(), path); err != nil {
		return httpResult{Error: err}
	} else {
		w.Write(out)
		return httpResult{}
	}
}
