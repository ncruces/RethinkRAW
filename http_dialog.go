package main

import (
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/ncruces/zenity"
)

var extensions = map[string]struct{}{}
var filters zenity.FileFilter

func init() {
	pattern := strings.Split("public.camera-raw-image"+
		"*.CRW *.NEF *.RAF *.ORF *.MRW *.DCR *.MOS *.RAW *.PEF *.SRF *.DNG *.X3F *.CR2 *.ERF *.SR2"+
		"*.KDC *.MFW *.MEF *.ARW *.NRW *.RW2 *.RWL *.IIQ *.3FR *.FFF *.SRW *.GPR *.DXO *.ARQ *.CR3",
		" ")
	filters = zenity.FileFilter{Name: "RAW photos", Patterns: pattern}
	for _, ext := range pattern[1:] {
		extensions[ext[1:]] = struct{}{}
	}
}

func dialogHandler(w http.ResponseWriter, r *http.Request) HTTPResult {
	if r.ParseForm() != nil {
		return HTTPResult{Status: http.StatusBadRequest}
	}

	var path string
	var paths []string

	_, photo := r.Form["photo"]
	_, batch := r.Form["batch"]
	_, gallery := r.Form["gallery"]

	switch {
	case batch:
		if res, err := zenity.SelectFileMutiple(zenity.Context(r.Context()), filters); err != nil {
			return HTTPResult{Error: err}
		} else {
			paths = res
		}

	case photo:
		if res, err := zenity.SelectFile(zenity.Context(r.Context()), filters); err != nil {
			return HTTPResult{Error: err}
		} else {
			path = res
		}

	case gallery:
		if res, err := zenity.SelectFile(zenity.Context(r.Context()), zenity.Directory()); err != nil {
			return HTTPResult{Error: err}
		} else {
			path = res
		}

	default:
		return HTTPResult{Status: http.StatusNotFound}
	}

	if path != "" {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return HTTPResult{Status: http.StatusUnprocessableEntity, Message: err.Error()}
		} else if err != nil {
			return HTTPResult{Error: err}
		}
	} else if len(paths) != 0 {
		path = batches.New(paths)
	} else {
		w.Header().Add("Refresh", "0; url=/")
		return HTTPResult{Status: http.StatusResetContent}
	}

	var url url.URL
	switch {
	case batch:
		url.Path = "/batch/" + path
	case photo:
		url.Path = "/photo/" + toURLPath(path)
	case gallery:
		url.Path = "/gallery/" + toURLPath(path)
	}
	return HTTPResult{
		Status:   http.StatusSeeOther,
		Location: url.String(),
	}
}
