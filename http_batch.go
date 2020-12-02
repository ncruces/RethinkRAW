package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"rethinkraw/osutil"

	"github.com/gorilla/schema"
	"github.com/ncruces/zenity"
)

type multiStatus struct {
	Code  int         `json:"code"`
	Text  string      `json:"text"`
	Body  interface{} `json:"response,omitempty"`
	Done  int         `json:"done,omitempty"`
	Total int         `json:"total,omitempty"`
}

func batchHandler(w http.ResponseWriter, r *http.Request) HTTPResult {
	if r.ParseForm() != nil {
		return HTTPResult{Status: http.StatusBadRequest}
	}

	// get batch
	id := strings.TrimPrefix(r.URL.Path, "/")
	batch := batches.Get(id)
	if len(batch) == 0 {
		path := fromURLPath(r.URL.Path)
		if fi, _ := os.Stat(path); fi != nil && fi.IsDir() {
			return HTTPResult{Location: "/batch/" + batches.New([]string{path})}
		}
		return HTTPResult{Status: http.StatusGone}
	}

	// get files in batch
	var files []string
	for _, path := range batch {
		err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if osutil.HiddenFile(info) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if info.Mode().IsRegular() {
				if _, ok := extensions[strings.ToUpper(filepath.Ext(path))]; ok {
					files = append(files, path)
				}
			}
			return nil
		})
		if err != nil {
			return HTTPResult{Error: err}
		}
	}

	_, save := r.Form["save"]
	_, export := r.Form["export"]
	_, settings := r.Form["settings"]

	switch {
	case save:
		var xmp xmpSettings
		dec := schema.NewDecoder()
		dec.IgnoreUnknownKeys(true)
		if err := dec.Decode(&xmp, r.Form); err != nil {
			return HTTPResult{Error: err}
		}
		xmp.Orientation = 0

		results := BatchProcessor(files, func(file string) error {
			xmp := xmp
			xmp.Filename = file
			return saveEdit(file, xmp)
		})

		w.Header().Set("Content-Type", "application/x-ndjson")
		w.WriteHeader(http.StatusMultiStatus)
		batchResultWriter(w, results, len(files))
		return HTTPResult{}

	case export:
		var xmp xmpSettings
		var exp exportSettings
		dec := schema.NewDecoder()
		dec.IgnoreUnknownKeys(true)
		if err := dec.Decode(&xmp, r.Form); err != nil {
			return HTTPResult{Error: err}
		}
		if err := dec.Decode(&exp, r.Form); err != nil {
			return HTTPResult{Error: err}
		}
		xmp.Orientation = 0

		exppath := filepath.Dir(files[0])
		if res, err := zenity.SelectFile(zenity.Directory(), zenity.Filename(exppath)); err != nil {
			return HTTPResult{Error: err}
		} else if res == "" {
			return HTTPResult{Status: http.StatusNoContent}
		} else {
			exppath = res
		}

		results := BatchProcessor(files, func(file string) error {
			var f io.WriteCloser
			xmp := xmp
			xmp.Filename = file
			out, err := exportEdit(file, xmp, exp)
			if err == nil {
				f, err = osutil.NewFile(filepath.Join(exppath, exportName(file, exp)))
			}
			if err == nil {
				_, err = f.Write(out)
			}
			if cerr := f.Close(); err == nil {
				err = cerr
			}
			return err
		})

		w.Header().Set("Content-Type", "application/x-ndjson")
		w.WriteHeader(http.StatusMultiStatus)
		batchResultWriter(w, results, len(files))
		return HTTPResult{}

	case settings:
		if xmp, err := loadEdit(files[0]); err != nil {
			return HTTPResult{Error: err}
		} else {
			w.Header().Set("Content-Type", "application/json")
			enc := json.NewEncoder(w)
			if err := enc.Encode(xmp); err != nil {
				return HTTPResult{Error: err}
			}
		}
		return HTTPResult{}

	default:
		w.Header().Set("Cache-Control", "max-age=300")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		data := struct {
			Photos []struct{ Name, Path string }
		}{}

		for _, file := range files {
			name := filepath.Base(file)
			item := struct{ Name, Path string }{name, toURLPath(file)}
			data.Photos = append(data.Photos, item)
		}

		return HTTPResult{
			Error: templates.ExecuteTemplate(w, "batch.gohtml", data),
		}
	}
}

func batchResultWriter(w http.ResponseWriter, results <-chan error, total int) {
	i := 0
	enc := json.NewEncoder(w)
	flush, _ := w.(http.Flusher)
	for err := range results {
		i += 1
		var status multiStatus
		if err != nil {
			status.Code, status.Body = errorStatus(err)
		} else {
			status.Code = http.StatusOK
		}
		status.Done, status.Total = i, total
		status.Text = http.StatusText(status.Code)
		enc.Encode(status)

		if flush != nil {
			flush.Flush()
		}
	}
}
