package main

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/ncruces/rethinkraw-pkg/osutil"
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
			Title, Path  string
			Dirs, Photos []struct{ Name, Path string }
		}{
			filepath.Clean(path),
			toURLPath(filepath.Clean(path)),
			nil, nil,
		}

		for _, info := range files {
			if osutil.HiddenFile(info) {
				continue
			}

			name := info.Name()
			path := filepath.Join(path, name)
			item := struct{ Name, Path string }{name, toURLPath(path)}

			if info.Mode()&os.ModeSymlink != 0 {
				info, err = os.Stat(path)
				if err != nil {
					continue
				}
			}
			const special = os.ModeType &^ os.ModeDir
			if info.Mode()&special != 0 {
				continue
			}

			if info.IsDir() {
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
