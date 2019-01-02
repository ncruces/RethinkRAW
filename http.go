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

func setupServer() *http.Server {
	templates = template.Must(template.ParseGlob("html/*.html"))
	http.Handle("/gallery/", http.StripPrefix("/gallery/", HttpHandler(galleryHandler)))
	http.Handle("/photo/", http.StripPrefix("/photo/", HttpHandler(photoHandler)))
	http.Handle("/thumb/", http.StripPrefix("/thumb/", HttpHandler(thumbHandler)))
	http.Handle("/", http.FileServer(http.Dir(filepath.Join(baseDir, "/static"))))
	return &http.Server{}
}

type HttpResult struct {
	Status   int
	Message  string
	Location string
}

type HttpHandler func(w http.ResponseWriter, r *http.Request) HttpResult

func (h HttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

func handleError(err error) HttpResult {
	if os.IsNotExist(err) {
		return HttpResult{Status: http.StatusNotFound}
	}

	if err, ok := err.(*exec.ExitError); ok {
		log.Println(string(err.Stderr))
	}

	log.Println(err)
	return HttpResult{Status: http.StatusInternalServerError}
}
