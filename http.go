package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
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

	if err, ok := err.(*exec.ExitError); ok {
		log.Println(string(err.Stderr))
	}

	log.Println(err)
	return HTTPResult{Status: http.StatusInternalServerError}
}
