package main

import (
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/ncruces/rethinkraw/pkg/osutil"
)

func galleryHandler(w http.ResponseWriter, r *http.Request) httpResult {
	if r := sendAllowed(w, r, "GET", "HEAD"); r.Done() {
		return r
	}
	prefix := getPathPrefix(r)
	path := fromURLPath(r.URL.Path, prefix)

	w.Header().Set("Cache-Control", "max-age=10")
	if r := sendCached(w, r, path); r.Done() {
		return r
	}

	if files, err := os.ReadDir(path); err != nil {
		return httpResult{Error: err}
	} else {
		data := struct {
			Title, Path  string
			Dirs, Photos []struct{ Name, Path string }
		}{
			filepath.Clean(path),
			toURLPath(path, prefix),
			nil, nil,
		}

		for _, entry := range files {
			name := entry.Name()
			path := filepath.Join(path, name)
			item := struct{ Name, Path string }{name, toURLPath(path, prefix)}

			if osutil.HiddenFile(entry) {
				continue
			}

			if entry.Type()&os.ModeSymlink != 0 {
				i, err := os.Stat(path)
				if err != nil {
					continue
				}
				entry = fs.FileInfoToDirEntry(i)
			}

			const special = os.ModeType &^ os.ModeDir
			if entry.Type()&special != 0 {
				continue
			}

			if entry.IsDir() {
				data.Dirs = append(data.Dirs, item)
			} else if _, ok := extensions[strings.ToUpper(filepath.Ext(name))]; ok {
				data.Photos = append(data.Photos, item)
			}
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		return httpResult{
			Error: templates.ExecuteTemplate(w, "gallery.gohtml", data),
		}
	}
}
