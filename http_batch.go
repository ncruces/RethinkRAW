package main

import (
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
)

func batchHandler(w http.ResponseWriter, r *http.Request) HTTPResult {
	path := fromURLPath(r.URL.Path)

	if files, err := ioutil.ReadDir(path); err != nil {
		return HTTPResult{Error: err}
	} else {
		w.Header().Set("Cache-Control", "max-age=10")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if r := cacheHeaders(path, r.Header, w.Header()); r.Done() {
			return r
		}

		data := struct {
			Title  string
			Photos []struct{ Name, Path string }
		}{}
		data.Title = filepath.Clean(path)

		for _, i := range files {
			if isHidden(i) {
				continue
			}

			name := i.Name()
			item := struct{ Name, Path string }{name, toURLPath(filepath.Join(path, name))}

			if _, ok := extensions[strings.ToUpper(filepath.Ext(name))]; ok && !i.IsDir() {
				data.Photos = append(data.Photos, item)
			}
		}

		return HTTPResult{
			Error: templates.ExecuteTemplate(w, "batch.gohtml", data),
		}
	}
}
