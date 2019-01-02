package main

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"
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

func galleryHandler(w http.ResponseWriter, r *http.Request) HttpResult {
	path := r.URL.Path

	if files, err := ioutil.ReadDir(path); err != nil {
		return handleError(err)
	} else {
		data := galleryData{
			Title:  filepath.Clean(path),
			Parent: filepath.ToSlash(filepath.Join(path, "..")),
		}

		for _, i := range files {
			if isHidden(i) {
				continue
			}

			name := i.Name()
			item := galleryItem{name, filepath.ToSlash(filepath.Join(path, name))}

			if i.IsDir() {
				data.Dirs = append(data.Dirs, item)
			} else if _, ok := extensions[strings.ToUpper(filepath.Ext(name))]; ok {
				data.Photos = append(data.Photos, item)
			}
		}
		w.Header().Add("Content-Type", "text/html")
		templates.ExecuteTemplate(w, "gallery.html", data)
		return HttpResult{}
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
