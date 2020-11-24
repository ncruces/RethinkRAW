package main

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"rethinkraw/osutil"
)

func galleryHandler(w http.ResponseWriter, r *http.Request) HTTPResult {
	path := fromURLPath(r.URL.Path)

	w.Header().Set("Cache-Control", "max-age=10")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if r := cacheHeaders(path, r.Header, w.Header()); r.Done() {
		return r
	}

	if files, err := ioutil.ReadDir(path); err != nil {
		return HTTPResult{Error: err}
	} else {
		data := struct {
			Title        string
			Dirs, Photos []struct{ Name, Path string }
		}{}
		data.Title = filepath.Clean(path)

		for _, i := range files {
			if osutil.HiddenFile(i) {
				continue
			}

			name := i.Name()
			path := filepath.Join(path, name)
			item := struct{ Name, Path string }{name, toURLPath(path)}

			if i.Mode()&os.ModeSymlink != 0 {
				i, err = os.Stat(path)
				if err != nil {
					continue
				}
			}
			const special = os.ModeType &^ os.ModeDir
			if i.Mode()&special != 0 {
				continue
			}

			if i.IsDir() {
				data.Dirs = append(data.Dirs, item)
			} else if _, ok := extensions[strings.ToUpper(filepath.Ext(name))]; ok {
				data.Photos = append(data.Photos, item)
			}
		}

		return HTTPResult{
			Error: templates.ExecuteTemplate(w, "gallery.gohtml", data),
		}
	}
}
