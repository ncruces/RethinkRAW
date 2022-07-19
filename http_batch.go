package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/schema"
	"github.com/ncruces/rethinkraw/pkg/osutil"
	"github.com/ncruces/zenity"
)

type multiStatus struct {
	Code  int    `json:"code"`
	Text  string `json:"text"`
	Body  any    `json:"response,omitempty"`
	Done  int    `json:"done,omitempty"`
	Total int    `json:"total,omitempty"`
}

func batchHandler(w http.ResponseWriter, r *http.Request) httpResult {
	if err := r.ParseForm(); err != nil {
		return httpResult{Status: http.StatusBadRequest, Error: err}
	}
	prefix := getPathPrefix(r)
	batch := fromBatchPath(r.URL.Path)
	if len(batch) == 0 {
		path := fromURLPath(r.URL.Path, prefix)
		if fi, _ := os.Stat(path); fi != nil && fi.IsDir() {
			return httpResult{Location: "/batch/" + toBatchPath(path)}
		}
		return httpResult{Status: http.StatusGone}
	}
	photos, err := findPhotos(batch)
	if err != nil {
		return httpResult{Error: err}
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
			return httpResult{Error: err}
		}
		xmp.Orientation = 0

		results := batchProcess(photos, func(photo batchPhoto) error {
			xmp := xmp
			xmp.Filename = filepath.Base(photo.Path)
			return saveEdit(photo.Path, xmp)
		})

		w.Header().Set("Content-Type", "application/x-ndjson")
		w.WriteHeader(http.StatusMultiStatus)
		batchResultWriter(w, results, len(photos))
		return httpResult{}

	case export:
		var xmp xmpSettings
		var exp exportSettings
		dec := schema.NewDecoder()
		dec.IgnoreUnknownKeys(true)
		if err := dec.Decode(&xmp, r.Form); err != nil {
			return httpResult{Error: err}
		}
		if err := dec.Decode(&exp, r.Form); err != nil {
			return httpResult{Error: err}
		}
		xmp.Orientation = 0

		var exppath string
		if len(photos) > 0 {
			exppath = filepath.Dir(photos[0].Path)
			if res, err := zenity.SelectFile(zenity.Context(r.Context()), zenity.Directory(), zenity.Filename(exppath)); res != "" {
				exppath = res
			} else if errors.Is(err, zenity.ErrCanceled) {
				return httpResult{Status: http.StatusNoContent}
			} else if err == nil {
				return httpResult{Status: http.StatusInternalServerError}
			} else {
				return httpResult{Error: err}
			}
		}

		results := batchProcess(photos, func(photo batchPhoto) (err error) {
			xmp := xmp
			xmp.Filename = filepath.Base(photo.Path)
			out, err := exportEdit(photo.Path, xmp, exp)
			if err != nil {
				return err
			}

			exppath := filepath.Join(exppath, exportPath(photo.Name, exp))
			if err := os.MkdirAll(filepath.Dir(exppath), 0777); err != nil {
				return err
			}
			f, err := osutil.NewFile(exppath)
			if err != nil {
				return err
			}
			defer func() {
				cerr := f.Close()
				if err == nil {
					err = cerr
				}
			}()

			_, err = f.Write(out)
			return err
		})

		w.Header().Set("Content-Type", "application/x-ndjson")
		w.WriteHeader(http.StatusMultiStatus)
		batchResultWriter(w, results, len(photos))
		return httpResult{}

	case settings:
		if len(photos) == 0 {
			return httpResult{Status: http.StatusNoContent}
		}
		if xmp, err := loadEdit(photos[0].Path); err != nil {
			return httpResult{Error: err}
		} else {
			w.Header().Set("Content-Type", "application/json")
			enc := json.NewEncoder(w)
			if err := enc.Encode(xmp); err != nil {
				return httpResult{Error: err}
			}
		}
		return httpResult{}

	default:
		w.Header().Set("Cache-Control", "max-age=10")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		data := struct {
			Photos []struct{ Name, Path string }
			Export bool
		}{
			nil, isLocalhost(r),
		}

		for _, photo := range photos {
			item := struct{ Name, Path string }{photo.Name, toURLPath(photo.Path, prefix)}
			data.Photos = append(data.Photos, item)
		}

		return httpResult{
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
