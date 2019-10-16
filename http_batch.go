package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func batchHandler(w http.ResponseWriter, r *http.Request) HTTPResult {
	path := fromURLPath(r.URL.Path)

	data := struct {
		Title  string
		Photos []struct{ Name, Path string }
	}{}
	data.Title = filepath.Clean(path)

	var walker = func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if isHidden(info) {
				return filepath.SkipDir
			}
		} else if _, ok := extensions[strings.ToUpper(filepath.Ext(info.Name()))]; ok {
			item := struct{ Name, Path string }{info.Name(), toURLPath(path)}
			data.Photos = append(data.Photos, item)
		}
		return nil
	}

	if err := filepath.Walk(path, walker); err != nil {
		return HTTPResult{Error: err}
	} else {
		w.Header().Set("Cache-Control", "max-age=10")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if r := cacheHeaders(path, r.Header, w.Header()); r.Done() {
			return r
		}

		return HTTPResult{
			Error: templates.ExecuteTemplate(w, "batch.gohtml", data),
		}
	}
}
