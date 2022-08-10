package main

import (
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/ncruces/jason"
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
			Template     jason.Object
		}{
			toUsrPath(path, prefix),
			toURLPath(path, prefix),
			nil, nil, jason.Object{},
		}
		if !isLocalhost(r) {
			data.Template["Upload"] = jason.Object{
				"Path": data.Path,
				"Exts": extensions,
			}
		}

		for _, entry := range files {
			name := entry.Name()
			path := filepath.Join(path, name)
			item := struct{ Name, Path string }{name, toURLPath(path, prefix)}

			if osutil.HiddenFile(entry) {
				continue
			}
			if entry.Type().IsRegular() {
				if _, ok := extensions[strings.ToUpper(filepath.Ext(name))]; ok {
					data.Photos = append(data.Photos, item)
				}
				continue
			}
			if entry.Type()&os.ModeSymlink != 0 {
				fi, err := os.Stat(path)
				if err != nil {
					continue
				}
				entry = fs.FileInfoToDirEntry(fi)
			}
			if entry.IsDir() {
				data.Dirs = append(data.Dirs, item)
			}
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		return httpResult{
			Error: templates.ExecuteTemplate(w, "gallery.gohtml", data),
		}
	}
}
