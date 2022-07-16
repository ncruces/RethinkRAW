package main

import (
	"errors"
	"net/http"
	"net/url"
	"os"

	"github.com/ncruces/zenity"
)

var extensions = map[string]struct{}{}
var filters zenity.FileFilter

func init() {
	pattern := []string{
		"public.camera-raw-image",
		"*.CRW", "*.NEF", "*.RAF", "*.ORF", "*.MRW", "*.DCR", "*.MOS", "*.RAW", "*.PEF", "*.SRF",
		"*.DNG", "*.X3F", "*.CR2", "*.ERF", "*.SR2", "*.KDC", "*.MFW", "*.MEF", "*.ARW", "*.NRW",
		"*.RW2", "*.RWL", "*.IIQ", "*.3FR", "*.FFF", "*.SRW", "*.GPR", "*.DXO", "*.ARQ", "*.CR3",
	}
	filters = zenity.FileFilter{Name: "RAW photos", Patterns: pattern}
	for _, ext := range pattern[1:] {
		extensions[ext[1:]] = struct{}{}
	}
}

func dialogHandler(w http.ResponseWriter, r *http.Request) httpResult {
	if err := r.ParseForm(); err != nil {
		return httpResult{Status: http.StatusBadRequest, Error: err}
	}
	if isLocal(r) {
		return httpResult{Status: http.StatusBadRequest}
	}

	var err error
	var path string
	var paths []string

	_, photo := r.Form["photo"]
	_, batch := r.Form["batch"]
	_, gallery := r.Form["gallery"]

	switch {
	case batch:
		paths, err = zenity.SelectFileMutiple(zenity.Context(r.Context()), filters)
	case photo:
		path, err = zenity.SelectFile(zenity.Context(r.Context()), filters)
	case gallery:
		path, err = zenity.SelectFile(zenity.Context(r.Context()), zenity.Directory())
	default:
		return httpResult{Status: http.StatusNotFound}
	}

	if path != "" {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return httpResult{Status: http.StatusUnprocessableEntity, Message: err.Error()}
		} else if err != nil {
			return httpResult{Error: err}
		}
	} else if len(paths) != 0 {
		path = toBatchPath(paths...)
	} else if errors.Is(err, zenity.ErrCanceled) {
		return httpResult{Status: http.StatusResetContent}
	} else if err == nil {
		return httpResult{Status: http.StatusInternalServerError}
	} else {
		return httpResult{Error: err}
	}

	var url url.URL
	switch {
	case batch:
		url.Path = "/batch/" + path
	case photo:
		url.Path = "/photo/" + toURLPath(path, "")
	case gallery:
		url.Path = "/gallery/" + toURLPath(path, "")
	}
	return httpResult{
		Status:   http.StatusSeeOther,
		Location: url.String(),
	}
}
