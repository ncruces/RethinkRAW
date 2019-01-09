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

func photoHandler(w http.ResponseWriter, r *http.Request) HTTPResult {
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
			return HTTPResult{}
		}

	case export:
		var xmp xmpSettings
		var exp exportSettings
		dec := schema.NewDecoder()
		dec.IgnoreUnknownKeys(true)
		if err := dec.Decode(&xmp, query); err != nil {
			return handleError(err)
		}
		if err := dec.Decode(&exp, query); err != nil {
			return handleError(err)
		}
		if out, err := exportEdit(path, &xmp, &exp); err != nil {
			return handleError(err)
		} else {
			w.Header().Add("Content-Disposition", "attachment")
			w.Header().Add("Content-Type", "image/jpeg")
			w.Write(out)
			return HTTPResult{}
		}

	case preview:
		var xmp xmpSettings
		dec := schema.NewDecoder()
		dec.IgnoreUnknownKeys(true)
		if err := dec.Decode(&xmp, query); err != nil {
			return handleError(err)
		}
		if out, err := previewEdit(path, &xmp); err != nil {
			return handleError(err)
		} else {
			w.Header().Add("Content-Type", "image/jpeg")
			w.Write(out)
			return HTTPResult{}
		}

	case settings:
		if xmp, err := loadEdit(path); err != nil {
			return handleError(err)
		} else {
			w.Header().Add("Content-Type", "application/json")
			enc := json.NewEncoder(w)
			if err := enc.Encode(xmp); err != nil {
				return handleError(err)
			}
		}
		return HTTPResult{}

	default:
		data := photoData{
			Name:   filepath.Base(path),
			Title:  filepath.Clean(path),
			Path:   filepath.ToSlash(filepath.Clean(path)),
			Parent: filepath.ToSlash(filepath.Join(path, "..")),
		}

		w.Header().Add("Content-Type", "text/html")
		return handleError(templates.ExecuteTemplate(w, "photo.html", data))
	}
}
