package main

import (
	"net/http"
	"os"
	"strings"
)

func batchHandler(w http.ResponseWriter, r *http.Request) HTTPResult {
	id := strings.TrimPrefix(r.URL.Path, "/")
	files := openMulti.get(id)
	r.ParseForm()

	if len(files) == 0 {
		return HTTPResult{Status: http.StatusGone}
	}

	w.Header().Set("Cache-Control", "max-age=300")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	data := struct {
		Photos []struct{ Name, Path string }
	}{}

	for _, f := range files {
		if fi, err := os.Stat(f); err != nil {
			return HTTPResult{Error: err}
		} else {
			name := fi.Name()
			item := struct{ Name, Path string }{name, toURLPath(f)}
			data.Photos = append(data.Photos, item)
		}
	}

	return HTTPResult{
		Error: templates.ExecuteTemplate(w, "batch.gohtml", data),
	}
}
