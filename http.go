package main

import (
	"bytes"
	"context"
	"encoding/json"
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

	"github.com/shurcooL/httpfs/html/vfstemplate"
	"github.com/shurcooL/httpgzip"
)

var templates *template.Template

func setupHTTP() *http.Server {
	templates = template.Must(vfstemplate.ParseGlob(assets, nil, "*.gohtml"))
	http.Handle("/gallery/", http.StripPrefix("/gallery/", HTTPHandler(galleryHandler)))
	http.Handle("/photo/", http.StripPrefix("/photo/", HTTPHandler(photoHandler)))
	http.Handle("/thumb/", http.StripPrefix("/thumb/", HTTPHandler(thumbHandler)))
	http.Handle("/config", HTTPHandler(configHandler))
	http.Handle("/", httpgzip.FileServer(assets, httpgzip.FileServerOptions{IndexHTML: true}))
	return &http.Server{}
}

// HTTPResult helps HTTPHandlers short circuit a result
type HTTPResult struct {
	Status   int
	Message  string
	Location string
	Error    error
}

func (r *HTTPResult) Done() bool { return r.Location != "" || r.Status != 0 || r.Error != nil }

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
		if res.Message == "" {
			res.Message = http.StatusText(res.Status)
		}
		http.Error(w, res.Message, res.Status)

	case res.Status != 0:
		w.WriteHeader(res.Status)

	case res.Error != nil:
		var status int
		var message strings.Builder

		switch {
		case os.IsNotExist(res.Error):
			status = http.StatusNotFound
		case os.IsPermission(res.Error):
			status = http.StatusForbidden
		default:
			status = http.StatusInternalServerError
		}

		message.WriteString(strings.TrimSpace(res.Error.Error()))
		if err, ok := res.Error.(*exec.ExitError); ok {
			if msg := bytes.TrimSpace(err.Stderr); len(msg) > 0 {
				message.WriteByte('\n')
				message.Write(msg)
			}
		}

		w.WriteHeader(status)

		if strings.HasPrefix(r.Header.Get("Accept"), "text/html") {
			templates.ExecuteTemplate(w, "error.gohtml", struct {
				Status, Message string
			}{
				http.StatusText(status),
				message.String(),
			})
		} else {
			json.NewEncoder(w).Encode(message.String())
		}

		log.Print(message.String())
	}
}

func cacheHeaders(path string, req, res http.Header) HTTPResult {
	if fi, err := os.Stat(path); err != nil {
		return HTTPResult{Error: err}
	} else {
		ims := req.Get("If-Modified-Since")
		if ims != "" {
			if t, err := http.ParseTime(ims); err == nil {
				if fi.ModTime().Before(t.Add(1 * time.Second)) {
					for k := range res {
						delete(res, k)
					}
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

	disposition := `attachment;filename="` + ascii + ext + `"`
	if ascii != utf {
		disposition += `;filename*=UTF-8''` + url.PathEscape(utf+ext)
	}

	headers.Set("Content-Disposition", disposition)
	headers.Set("Content-Type", mime.TypeByExtension(ext))
}

func toURLPath(path string) string {
	if strings.HasPrefix(path, `\\`) {
		return `\\` + filepath.ToSlash(path[2:])
	}
	return filepath.ToSlash(path)
}
