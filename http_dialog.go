package main

import (
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/ncruces/zenity"
)

var filters zenity.FileFilters
var extensions = map[string]struct{}{}

func init() {
	pattern := strings.Split("*.CRW *.NEF *.RAF *.ORF *.MRW *.DCR *.MOS *.RAW *.PEF *.SRF *.DNG *.X3F *.CR2 *.ERF *.SR2 *.KDC *.MFW *.MEF *.ARW *.NRW *.RW2 *.RWL *.IIQ *.3FR *.FFF *.SRW *.GPR *.DXO *.ARQ *.CR3", " ")
	filters = zenity.FileFilters{{Name: "RAW files", Patterns: pattern}}
	for _, ext := range pattern {
		extensions[ext[1:]] = struct{}{}
	}
}

func dialogHandler(w http.ResponseWriter, r *http.Request) HTTPResult {
	if r.ParseForm() != nil {
		return HTTPResult{Status: http.StatusBadRequest}
	}

	var path string
	var paths []string

	nonce := r.Form.Get("nonce")
	_, photo := r.Form["photo"]
	_, batch := r.Form["batch"]
	_, gallery := r.Form["gallery"]

	bringToTop()
	switch {
	case batch:
		if res, err := zenity.SelectFileMutiple(filters.New()); err != nil {
			return HTTPResult{Error: err}
		} else {
			paths = res
		}

	case photo:
		if res, err := zenity.SelectFile(filters.New()); err != nil {
			return HTTPResult{Error: err}
		} else {
			path = res
		}

	case gallery:
		if res, err := zenity.SelectFile(zenity.Directory); err != nil {
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
		openMulti.put(nonce, paths)
	} else {
		w.Header().Add("Refresh", "0; url=/")
		return HTTPResult{Status: http.StatusResetContent}
	}

	var url url.URL
	switch {
	case batch:
		url.Path = "/batch/" + nonce
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

var openMulti OpenMulti

type OpenMulti struct {
	lock  sync.Mutex
	queue [16]struct {
		id    string
		paths []string
	}
}

func (p *OpenMulti) put(id string, paths []string) {
	p.lock.Lock()
	defer p.lock.Unlock()

	for i := len(p.queue) - 1; i > 0; i-- {
		p.queue[i] = p.queue[i-1]
	}

	p.queue[0].id = id
	p.queue[0].paths = paths
}

func (p *OpenMulti) get(id string) []string {
	p.lock.Lock()
	defer p.lock.Unlock()

	for j, t := range p.queue {
		if t.id == id {
			for i := j; i > 0; i-- {
				p.queue[i] = p.queue[i-1]
			}
			p.queue[0] = t
			return t.paths
		}
	}

	return nil
}
