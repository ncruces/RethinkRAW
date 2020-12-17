package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/schema"
	"github.com/ncruces/rethinkraw/pkg/osutil"
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
	photos, err := FindPhotos(batch)
	if err != nil {
		return HTTPResult{Error: err}
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

		results := BatchProcess(photos, func(photo BatchPhoto) error {
			xmp := xmp
			xmp.Filename = photo.Path
			return saveEdit(photo.Path, xmp)
		})

		w.Header().Set("Content-Type", "application/x-ndjson")
		w.WriteHeader(http.StatusMultiStatus)
		batchResultWriter(w, results, len(photos))
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

		var exppath string
		if len(photos) > 0 {
			exppath = filepath.Dir(photos[0].Path)
			if res, err := zenity.SelectFile(zenity.Directory(), zenity.Filename(exppath)); err != nil {
				return HTTPResult{Error: err}
			} else if res == "" {
				return HTTPResult{Status: http.StatusNoContent}
			} else {
				exppath = res
			}
		}

		results := BatchProcess(photos, func(photo BatchPhoto) (err error) {
			xmp := xmp
			xmp.Filename = photo.Path
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
		return HTTPResult{}

	case settings:
		if len(photos) == 0 {
			return HTTPResult{Status: http.StatusNoContent}
		}
		if xmp, err := loadEdit(photos[0].Path); err != nil {
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

		for _, photo := range photos {
			item := struct{ Name, Path string }{photo.Name, toURLPath(photo.Path)}
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
