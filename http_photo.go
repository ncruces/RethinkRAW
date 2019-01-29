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
	r.ParseForm()

	_, meta := r.Form["meta"]
	_, save := r.Form["save"]
	_, export := r.Form["export"]
	_, preview := r.Form["preview"]
	_, settings := r.Form["settings"]

	switch {
	case meta:
		res := cacheHeaders(path, r.Header, w.Header())
		if res.Status == 0 {
			if out, err := getMeta(path); err != nil {
				return handleError(err)
			} else {
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				w.Write(out)
			}
		}
		return res

	case save:
		var xmp xmpSettings
		xmp.Filename = path
		dec := schema.NewDecoder()
		dec.IgnoreUnknownKeys(true)
		if err := dec.Decode(&xmp, r.Form); err != nil {
			return handleError(err)
		}
		if err := saveEdit(path, &xmp); err != nil {
			return handleError(err)
		} else {
			return HTTPResult{Status: http.StatusNoContent}
		}

	case export:
		var xmp xmpSettings
		xmp.Filename = path
		var exp exportSettings
		dec := schema.NewDecoder()
		dec.IgnoreUnknownKeys(true)
		if err := dec.Decode(&xmp, r.Form); err != nil {
			return handleError(err)
		}
		if err := dec.Decode(&exp, r.Form); err != nil {
			return handleError(err)
		}
		if out, err := exportEdit(path, &xmp, &exp); err != nil {
			return handleError(err)
		} else {
			exportHeaders(path, &exp, w.Header())
			w.Write(out)
			return HTTPResult{}
		}

	case preview:
		var xmp xmpSettings
		dec := schema.NewDecoder()
		dec.IgnoreUnknownKeys(true)
		if err := dec.Decode(&xmp, r.Form); err != nil {
			return handleError(err)
		}
		if out, err := previewEdit(path, &xmp); err != nil {
			return handleError(err)
		} else {
			w.Header().Set("Content-Type", "image/jpeg")
			w.Write(out)
			return HTTPResult{}
		}

	case settings:
		if xmp, err := loadEdit(path); err != nil {
			return handleError(err)
		} else {
			w.Header().Set("Content-Type", "application/json")
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
			Path:   toURLPath(filepath.Clean(path)),
			Parent: toURLPath(filepath.Join(path, "..")),
		}

		w.Header().Set("Content-Type", "text/html")
		return handleError(templates.ExecuteTemplate(w, "photo.html", data))
	}
}
