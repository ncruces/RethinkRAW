package main

import (
	"net/http"
	"net/url"
	"os"
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

var filter = "CRW,NEF,RAF,ORF,MRW,DCR,MOS,RAW,PEF,SRF,DNG,X3F,CR2,ERF,SR2,KDC,MFW,MEF,ARW,NRW,RW2,RWL,IIQ,3FR,FFF,SRW,GPR,DXO,ARQ,CR3"

func browseHandler(w http.ResponseWriter, r *http.Request) HTTPResult {
	var path string
	r.ParseForm()

	bringToTop()
	if _, photo := r.Form["photo"]; photo {
		f := strings.Join(append([]string{filter}, strings.Split(filter, ",")...), ";")
		if res, err := nfd.OpenDialog(f, path); err != nil {
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
		w.Header().Add("Refresh", "0; url=/")
		return HTTPResult{Status: http.StatusResetContent}
	} else if fi, err := os.Stat(path); os.IsNotExist(err) {
		w.Header().Add("Refresh", "0; url=/")
		return HTTPResult{Status: http.StatusResetContent}
	} else if err != nil {
		return HTTPResult{Error: err}
	} else {
		var url url.URL
		if fi.IsDir() {
			if _, gallery := r.Form["gallery"]; gallery {
				url.Path = "/gallery/" + toURLPath(path)
			} else {
				url.Path = "/batch/" + toURLPath(path)
			}
		} else {
			url.Path = "/photo/" + toURLPath(path)
		}
		return HTTPResult{
			Status:   http.StatusSeeOther,
			Location: url.String(),
		}
	}
}
