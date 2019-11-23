package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/schema"
)

func batchHandler(w http.ResponseWriter, r *http.Request) HTTPResult {
	id := strings.TrimPrefix(r.URL.Path, "/")
	files := openMulti.get(id)
	r.ParseForm()

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

		for _, path := range files {
			if err := saveEdit(path, &xmp); err != nil {
				return HTTPResult{Error: err}
			}
		}
		return HTTPResult{Status: http.StatusNoContent}

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

		log.Printf("Exporting: %+v; %+v", xmp, exp)
		return HTTPResult{Status: http.StatusNoContent}

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
