package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/schema"
	"github.com/ncruces/zenity"
)

func photoHandler(w http.ResponseWriter, r *http.Request) httpResult {
	if err := r.ParseForm(); err != nil {
		return httpResult{Status: http.StatusBadRequest, Error: err}
	}
	prefix := getPathPrefix(r)
	path := fromURLPath(r.URL.Path, prefix)

	_, meta := r.Form["meta"]
	_, save := r.Form["save"]
	_, export := r.Form["export"]
	_, preview := r.Form["preview"]
	_, settings := r.Form["settings"]
	_, whiteBalance := r.Form["wb"]
	_, print := r.Form["print"]

	switch {
	case meta:
		w.Header().Set("Cache-Control", "max-age=60")
		if r := sendCached(w, r, path); r.Done() {
			return r
		}

		if out, err := getMetaHTML(path); err != nil {
			return httpResult{Error: err}
		} else {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(out)
			return httpResult{}
		}

	case save:
		var xmp xmpSettings
		dec := schema.NewDecoder()
		dec.IgnoreUnknownKeys(true)
		if err := dec.Decode(&xmp, r.Form); err != nil {
			return httpResult{Error: err}
		}
		xmp.Filename = path

		if err := saveEdit(path, xmp); err != nil {
			return httpResult{Error: err}
		} else {
			return httpResult{Status: http.StatusNoContent}
		}

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
		xmp.Filename = path

		exppath := exportPath(path, exp)
		if res, err := zenity.SelectFileSave(zenity.Context(r.Context()), zenity.Filename(exppath), zenity.ConfirmOverwrite()); res != "" {
			exppath = res
		} else if errors.Is(err, zenity.ErrCanceled) {
			return httpResult{Status: http.StatusNoContent}
		} else if err == nil {
			return httpResult{Status: http.StatusInternalServerError}
		} else {
			return httpResult{Error: err}
		}

		if out, err := exportEdit(path, xmp, exp); err != nil {
			return httpResult{Error: err}
		} else if err := ioutil.WriteFile(exppath, out, 0666); err != nil {
			return httpResult{Error: err}
		} else {
			return httpResult{Status: http.StatusNoContent}
		}

	case preview:
		var xmp xmpSettings
		var size struct{ Preview int }
		dec := schema.NewDecoder()
		dec.IgnoreUnknownKeys(true)
		if err := dec.Decode(&xmp, r.Form); err != nil {
			return httpResult{Error: err}
		}
		if err := dec.Decode(&size, r.Form); err != nil {
			return httpResult{Error: err}
		}
		if out, err := previewEdit(path, size.Preview, xmp); err != nil {
			return httpResult{Error: err}
		} else {
			w.Header().Set("Content-Type", "image/jpeg")
			w.Write(out)
			return httpResult{}
		}

	case settings:
		if xmp, err := loadEdit(path); err != nil {
			return httpResult{Error: err}
		} else {
			w.Header().Set("Content-Type", "application/json")
			enc := json.NewEncoder(w)
			if err := enc.Encode(xmp); err != nil {
				return httpResult{Error: err}
			}
		}
		return httpResult{}

	case whiteBalance:
		var coords struct{ WB []float64 }
		dec := schema.NewDecoder()
		dec.IgnoreUnknownKeys(true)
		if err := dec.Decode(&coords, r.Form); err != nil {
			return httpResult{Error: err}
		}
		if wb, err := loadWhiteBalance(path, coords.WB); err != nil {
			return httpResult{Error: err}
		} else {
			w.Header().Set("Content-Type", "application/json")
			enc := json.NewEncoder(w)
			if err := enc.Encode(wb); err != nil {
				return httpResult{Error: err}
			}
		}
		return httpResult{}

	case print:
		if _, err := os.Stat(path); err != nil {
			return httpResult{Error: err}
		}

		r.Form.Del("print")
		r.Form.Set("preview", "")

		w.Header().Set("Cache-Control", "max-age=300")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		return httpResult{
			Error: templates.ExecuteTemplate(w, "print.gohtml", "?"+r.Form.Encode()),
		}

	default:
		if _, err := os.Stat(path); err != nil {
			return httpResult{Error: err}
		}

		w.Header().Set("Cache-Control", "max-age=300")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		return httpResult{
			Error: templates.ExecuteTemplate(w, "photo.gohtml", struct {
				Name, Title, Path string
			}{
				filepath.Base(path),
				filepath.Clean(path),
				toURLPath(path, prefix),
			}),
		}
	}
}
