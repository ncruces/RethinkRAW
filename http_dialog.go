package main

import (
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

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

func dialogHandler(w http.ResponseWriter, r *http.Request) HTTPResult {
	var paths []string
	var path string
	r.ParseForm()

	nonce := r.Form.Get("nonce")
	_, photo := r.Form["photo"]
	_, batch := r.Form["batch"]
	_, gallery := r.Form["gallery"]

	bringToTop()
	switch {
	case batch:
		f := strings.Join(append([]string{filter}, strings.Split(filter, ",")...), ";")
		if res, err := nfd.OpenDialogMultiple(f, path); err != nil {
			return HTTPResult{Error: err}
		} else {
			paths = res
		}

	case photo:
		f := strings.Join(append([]string{filter}, strings.Split(filter, ",")...), ";")
		if res, err := nfd.OpenDialog(f, path); err != nil {
			return HTTPResult{Error: err}
		} else {
			path = res
		}

	case gallery:
		if res, err := nfd.PickFolder(path); err != nil {
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
	queue [4]struct {
		id    string
		paths []string
	}
}

func (p *OpenMulti) put(id string, paths []string) {
	p.lock.Lock()
	defer p.lock.Unlock()

	type multiPath struct {
		id    string
		paths []string
	}

	for i := len(p.queue) - 1; i > 0; i-- {
		p.queue[i] = p.queue[i-1]
	}
	p.queue[0] = multiPath{id, paths}
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
