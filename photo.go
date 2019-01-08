package main

import (
	"encoding/json"
	"net/http"
	"path/filepath"

	"github.com/gorilla/schema"
)

type photoData struct {
	Title, Parent, Name, Path string
}

func photoHandler(w http.ResponseWriter, r *http.Request) HttpResult {
	path := r.URL.Path
	query := r.URL.Query()

	_, meta := query["meta"]
	_, export := query["export"]
	_, preview := query["preview"]
	_, settings := query["settings"]

	switch {
	case meta:
		if out, err := getMeta(path); err != nil {
			return handleError(err)
		} else {
			w.Header().Add("Content-Type", "text/plain")
			w.Write(out)
			return HttpResult{}
		}

	case export:
		var set xmpSettings
		dec := schema.NewDecoder()
		dec.IgnoreUnknownKeys(true)
		if err := dec.Decode(&set, query); err != nil {
			return handleError(err)
		}
		if out, err := exportEdit(path, &set); err != nil {
			return handleError(err)
		} else {
			w.Header().Add("Content-Disposition", "attachment")
			w.Header().Add("Content-Type", "image/jpeg")
			w.Write(out)
			return HttpResult{}
		}

	case preview:
		var set xmpSettings
		dec := schema.NewDecoder()
		dec.IgnoreUnknownKeys(true)
		if err := dec.Decode(&set, query); err != nil {
			return handleError(err)
		}
		if out, err := previewEdit(path, &set); err != nil {
			return handleError(err)
		} else {
			w.Header().Add("Content-Type", "image/jpeg")
			w.Write(out)
			return HttpResult{}
		}

	case settings:
		if set, err := getEditSettings(path); err != nil {
			return handleError(err)
		} else {
			w.Header().Add("Content-Type", "application/json")
			enc := json.NewEncoder(w)
			if err := enc.Encode(set); err != nil {
				return handleError(err)
			}
		}
		return HttpResult{}

	default:
		data := photoData{
			Name:   filepath.Base(path),
			Title:  filepath.Clean(path),
			Path:   filepath.ToSlash(filepath.Clean(path)),
			Parent: filepath.ToSlash(filepath.Join(path, "..")),
		}

		w.Header().Add("Content-Type", "text/html")
		templates.ExecuteTemplate(w, "photo.html", data)
		return HttpResult{}
	}
}
