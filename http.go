package main

import (
	"context"
	"html/template"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var templates *template.Template

func setupHTTP() *http.Server {
	templates = template.Must(template.ParseGlob("html/*.html"))
	http.Handle("/gallery/", http.StripPrefix("/gallery/", HTTPHandler(galleryHandler)))
	http.Handle("/photo/", http.StripPrefix("/photo/", HTTPHandler(photoHandler)))
	http.Handle("/thumb/", http.StripPrefix("/thumb/", HTTPHandler(thumbHandler)))
	http.Handle("/", http.FileServer(http.Dir(filepath.Join(baseDir, "/static"))))
	return &http.Server{}
}

// HTTPResult helps HTTPHandlers short circuit a result
type HTTPResult struct {
	Status   int
	Message  string
	Location string
}

// HTTPHandler is an http.Handler that returns an HTTPResult
type HTTPHandler func(w http.ResponseWriter, r *http.Request) HTTPResult

func (h HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	switch res := h(w, r.WithContext(ctx)); {

	case res.Location != "":
		if res.Status == 0 {
			res.Status = http.StatusTemporaryRedirect
		}
		http.Redirect(w, r, res.Location, res.Status)

	case res.Status >= 400:
		h := w.Header()
		for k := range h {
			delete(h, k)
		}
		if res.Message == "" {
			res.Message = http.StatusText(res.Status)
		}
		http.Error(w, res.Message, res.Status)

	case res.Status != 0:
		w.WriteHeader(res.Status)

	}
}

func handleError(err error) HTTPResult {
	if err == nil {
		return HTTPResult{}
	}

	if os.IsNotExist(err) {
		return HTTPResult{Status: http.StatusNotFound}
	}

	if os.IsPermission(err) {
		return HTTPResult{Status: http.StatusForbidden}
	}

	if err, ok := err.(*exec.ExitError); ok {
		log.Println(string(err.Stderr))
	}

	log.Println(err)
	return HTTPResult{Status: http.StatusInternalServerError}
}

func cacheHeaders(path string, req, res http.Header) HTTPResult {
	if fi, err := os.Stat(path); err != nil {
		return handleError(err)
	} else {
		ims := req.Get("If-Modified-Since")
		if ims != "" {
			if t, err := http.ParseTime(ims); err == nil {
				if fi.ModTime().Before(t.Add(1 * time.Second)) {
					return HTTPResult{Status: http.StatusNotModified}
				}
			}
		}
		res.Set("Last-Modified", fi.ModTime().UTC().Format(http.TimeFormat))
	}
	return HTTPResult{}
}

func attachmentHeaders(path, ext string, headers http.Header) {
	if ext == "" {
		ext = filepath.Ext(path)
	}
	path = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))

	utf := filename(path)
	ascii := filename(toASCII(path))
	if utf == "" {
		utf = "download"
	}
	if ascii == "" {
		ascii = "download"
	}

	headers.Set("Content-Type", mime.TypeByExtension(ext))
	headers.Set("Content-Disposition", `attachment; filename="`+ascii+ext+`"; filename*=UTF-8''`+url.PathEscape(utf+ext))
}

func toURLPath(path string) string {
	if strings.HasPrefix(path, `\\`) {
		return `\\` + filepath.ToSlash(path[2:])
	}
	return filepath.ToSlash(path)
}
