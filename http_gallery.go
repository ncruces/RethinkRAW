package main

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	nfd "github.com/ncruces/go-nativefiledialog"
)

var extensions = map[string]struct{}{
	".CRW": {}, // Canon
	".NEF": {}, // Nikon
	".RAF": {}, // Fujifilm
	".ORF": {}, // Olympus
	".MRW": {}, // Minolta
	".DCR": {}, // Kodak
	".MOS": {}, // Leaf
	".RAW": {}, // Panasonic
	".PEF": {}, // Pentax
	".SRF": {}, // Sony
	".DNG": {}, // Adobe
	".X3F": {}, // Sigma
	".CR2": {}, // Canon
	".ERF": {}, // Epson
	".SR2": {}, // Sony
	".KDC": {}, // Kodak
	".MFW": {}, // Mamiya
	".MEF": {}, // Mamiya
	".ARW": {}, // Sony
	".NRW": {}, // Nikon
	".RW2": {}, // Panasonic
	".RWL": {}, // Leica
	".IIQ": {}, // Phase One
	".3FR": {}, // Hasselblad
	".FFF": {}, // Hasselblad
	".SRW": {}, // Samsung
	".GPR": {}, // GoPro
	".DXO": {}, // DxO
	".ARQ": {}, // Sony
	".CR3": {}, // Canon
}

func galleryHandler(w http.ResponseWriter, r *http.Request) HTTPResult {
	path := r.URL.Path
	r.ParseForm()

	_, edit := r.Form["edit"]
	_, browse := r.Form["browse"]
	if edit || browse {
		bringToTop()
		if edit {
			exts := make([]string, 0, len(extensions))
			for e := range extensions {
				exts = append(exts, e[1:])
			}
			sort.Strings(exts)

			if res, err := nfd.OpenDialog(strings.Join(exts, ","), path); err != nil {
				return HTTPResult{Error: err}
			} else {
				path = res
			}
		} else {
			if res, err := nfd.PickFolder(path); err != nil {
				return HTTPResult{Error: err}
			} else {
				path = res
			}
		}

		if path == "" {
			return HTTPResult{Status: http.StatusResetContent}
		} else if fi, err := os.Stat(path); os.IsNotExist(err) {
			return HTTPResult{Status: http.StatusResetContent}
		} else if err != nil {
			return HTTPResult{Error: err}
		} else {
			var url url.URL
			if fi.IsDir() {
				url.Path = "/gallery/" + toURLPath(path)
			} else {
				url.Path = "/photo/" + toURLPath(path)
			}
			return HTTPResult{
				Status:   http.StatusSeeOther,
				Location: url.String(),
			}
		}
	}

	if r := cacheHeaders(path, r.Header, w.Header()); r.Done() {
		return r
	}

	if files, err := ioutil.ReadDir(path); err != nil {
		return HTTPResult{Error: err}
	} else {
		path = filepath.Join(path, ".")

		data := struct {
			Title        string
			Dirs, Photos []struct{ Name, Path string }
		}{}
		data.Title = filepath.Clean(path)

		for _, i := range files {
			if isHidden(i) {
				continue
			}

			name := i.Name()
			item := struct{ Name, Path string }{name, toURLPath(filepath.Join(path, name))}

			if i.IsDir() {
				data.Dirs = append(data.Dirs, item)
			} else if _, ok := extensions[strings.ToUpper(filepath.Ext(name))]; ok {
				data.Photos = append(data.Photos, item)
			}
		}

		w.Header().Set("Cache-Control", "max-age=60")
		w.Header().Set("Content-Type", "text/html")
		return HTTPResult{
			Error: templates.ExecuteTemplate(w, "gallery.html", data),
		}
	}
}
