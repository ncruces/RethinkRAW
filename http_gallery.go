package main

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	nfd "./go-nfd"
)

type galleryItem struct {
	Name, Path string
}

type galleryData struct {
	Title, Parent string
	Dirs, Photos  []galleryItem
}

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
	query := r.URL.Query()

	_, browse := query["browse"]
	if browse {
		bringToTop()
		if folder, err := nfd.PickFolder(r.URL.Path); err != nil {
			return handleError(err)
		} else if folder == "" {
			return HTTPResult{Status: http.StatusResetContent}
		} else {
			url := url.URL{Path: "/gallery/" + toURLPath(folder)}
			return HTTPResult{
				Status:   http.StatusSeeOther,
				Location: url.String(),
			}
		}
	}

	if files, err := ioutil.ReadDir(path); err != nil {
		return handleError(err)
	} else {
		path = filepath.Join(path, ".")
		parent := filepath.Join(path, "..")
		if path == parent {
			parent = ""
		}

		data := galleryData{
			Title:  filepath.Clean(path),
			Parent: toURLPath(parent),
		}

		for _, i := range files {
			if isHidden(i) {
				continue
			}

			name := i.Name()
			item := galleryItem{name, toURLPath(filepath.Join(path, name))}

			if i.IsDir() {
				data.Dirs = append(data.Dirs, item)
			} else if _, ok := extensions[strings.ToUpper(filepath.Ext(name))]; ok {
				data.Photos = append(data.Photos, item)
			}
		}
		w.Header().Set("Content-Type", "text/html")
		return handleError(templates.ExecuteTemplate(w, "gallery.html", data))
	}
}

func isHidden(fi os.FileInfo) bool {
	if strings.HasPrefix(fi.Name(), ".") {
		return true
	}

	if s, ok := fi.Sys().(*syscall.Win32FileAttributeData); ok &&
		s.FileAttributes&(syscall.FILE_ATTRIBUTE_HIDDEN|syscall.FILE_ATTRIBUTE_SYSTEM) != 0 {
		return true
	}

	return false
}
