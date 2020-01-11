package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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

	id := strings.TrimPrefix(r.URL.Path, "/")
	files := openMulti.get(id)
	if len(files) == 0 {
		return HTTPResult{Status: http.StatusGone}
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

		w.Header().Set("Content-Type", "application/x-ndjson")
		w.WriteHeader(http.StatusMultiStatus)

		enc := json.NewEncoder(w)
		flush, _ := w.(http.Flusher)
		for i, file := range files {
			var status multiStatus
			err := saveEdit(file, &xmp)
			if err != nil {
				status.Code, status.Body = errorStatus(err)
			} else {
				status.Code = http.StatusOK
			}
			status.Done, status.Total = i+1, len(files)
			status.Text = http.StatusText(status.Code)
			enc.Encode(status)

			if flush != nil {
				flush.Flush()
			}
		}
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

		path := filepath.Dir(files[1])
		if res, err := zenity.SelectFile(zenity.Directory, zenity.Filename(path)); err != nil {
			return HTTPResult{Error: err}
		} else if res == "" {
			return HTTPResult{Status: http.StatusNoContent}
		} else {
			path = res
		}

		w.Header().Set("Content-Type", "application/x-ndjson")
		w.WriteHeader(http.StatusMultiStatus)

		enc := json.NewEncoder(w)
		flush, _ := w.(http.Flusher)
		for i, file := range files {
			var n int
			var f io.WriteCloser
			var status multiStatus
			out, err := exportEdit(file, &xmp, &exp)
			if err == nil {
				f, err = makeFile(filepath.Join(path, exportName(file, &exp)))
			}
			if err == nil {
				n, err = f.Write(out)
			}
			if err == nil && n < len(out) {
				err = io.ErrShortWrite
			}
			if cerr := f.Close(); err == nil {
				err = cerr
			}
			if err == nil {
				status.Code = http.StatusOK
			} else {
				status.Code, status.Body = errorStatus(err)
			}
			status.Done, status.Total = i+1, len(files)
			status.Text = http.StatusText(status.Code)
			enc.Encode(status)

			if flush != nil {
				flush.Flush()
			}
		}
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
}
