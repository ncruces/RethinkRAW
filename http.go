package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"html/template"
	"log"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"golang.org/x/net/websocket"
)

var templates *template.Template

func setupHTTP() *http.Server {
	http.Handle("/gallery/", http.StripPrefix("/gallery", HTTPHandler(galleryHandler)))
	http.Handle("/photo/", http.StripPrefix("/photo", HTTPHandler(photoHandler)))
	http.Handle("/batch/", http.StripPrefix("/batch", HTTPHandler(batchHandler)))
	http.Handle("/thumb/", http.StripPrefix("/thumb", HTTPHandler(thumbHandler)))
	http.Handle("/dialog", HTTPHandler(dialogHandler))
	http.Handle("/ws", websocket.Handler(websocketWatcher))
	http.Handle("/", assetHandler)
	templates = assetTemplates()
	return &http.Server{
		ReadHeaderTimeout: time.Second,
		IdleTimeout:       time.Minute,
		ConnState:         connectionWatcher,
		Handler:           middlewareWatcher(http.DefaultServeMux),
	}
}

var watchdog struct {
	*time.Ticker
	numActive  int32
	lastActive int64
}

func init() {
	watchdog.Ticker = time.NewTicker(time.Minute)
	watchdog.lastActive = time.Now().Add(time.Hour).UnixNano()

	go func() {
		for range watchdog.C {
			if atomic.LoadInt32(&watchdog.numActive) > 0 {
				continue
			}
			t := time.Unix(0, atomic.LoadInt64(&watchdog.lastActive))
			if time.Now().After(t.Add(time.Minute)) {
				shutdown <- os.Interrupt
			}
		}
	}()
}

func middlewareWatcher(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&watchdog.numActive, +1)
		defer atomic.AddInt32(&watchdog.numActive, -1)
		atomic.StoreInt64(&watchdog.lastActive, time.Now().UnixNano())
		next.ServeHTTP(w, r)
	})
}

func connectionWatcher(conn net.Conn, state http.ConnState) {
	switch state {
	case http.StateNew:
		atomic.AddInt32(&watchdog.numActive, +1)
	case http.StateHijacked, http.StateClosed:
		atomic.AddInt32(&watchdog.numActive, -1)
	case http.StateActive:
		atomic.StoreInt64(&watchdog.lastActive, time.Now().UnixNano())
	}
}

func websocketWatcher(conn *websocket.Conn) {
	conn.PayloadType = websocket.PingFrame

	for {
		var dummy [8]byte
		conn.SetReadDeadline(time.Now().Add(time.Minute))
		if _, err := conn.Read(dummy[:]); err == nil {
			continue
		} else if !os.IsTimeout(err) {
			return
		}

		conn.SetWriteDeadline(time.Now().Add(time.Second))
		if _, err := conn.Write(nil); err != nil {
			return
		}
	}
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
		sendError(w, r, res.Status, res.Message)

	case res.Status != 0:
		w.WriteHeader(res.Status)

	case res.Error != nil:
		status, message := errorStatus(res.Error)
		sendError(w, r, status, message)
		log.Print(message)
	}
}

func errorStatus(err error) (status int, message string) {
	switch {
	case os.IsNotExist(err):
		status = http.StatusNotFound
	case os.IsPermission(err):
		status = http.StatusForbidden
	default:
		status = http.StatusInternalServerError
	}

	var buf strings.Builder
	buf.WriteString(strings.TrimSpace(err.Error()))

	var eerr *exec.ExitError
	if errors.As(err, &eerr) {
		if msg := bytes.TrimSpace(eerr.Stderr); len(msg) > 0 {
			buf.WriteByte('\n')
			buf.Write(msg)
		}
	}

	return status, buf.String()
}

func sendError(w http.ResponseWriter, r *http.Request, status int, message string) {
	h := w.Header()
	for n := range h {
		delete(h, n)
	}
	h.Set("X-Content-Type-Options", "nosniff")
	if strings.HasPrefix(r.Header.Get("Accept"), "text/html") {
		h.Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(status)
		templates.ExecuteTemplate(w, "error.gohtml", struct {
			Status, Message string
		}{
			http.StatusText(status),
			message,
		})
	} else {
		h.Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(message)
	}
}

func cacheHeaders(path string, req, res http.Header) HTTPResult {
	if fi, err := os.Stat(path); err != nil {
		return HTTPResult{Error: err}
	} else {
		res.Set("Last-Modified", fi.ModTime().UTC().Format(http.TimeFormat))
		if ims := req.Get("If-Modified-Since"); ims != "" {
			if t, err := http.ParseTime(ims); err == nil {
				if fi.ModTime().Before(t.Add(1 * time.Second)) {
					for k := range res {
						switch k {
						case "Cache-Control", "Last-Modified":
							// keep
						default:
							delete(res, k)
						}
					}
					return HTTPResult{Status: http.StatusNotModified}
				}
			}
		}
	}
	return HTTPResult{}
}

func attachmentHeaders(path string, headers http.Header) {
	ext := filepath.Ext(path)
	utf := filename(path)
	ascii := filename(toASCII(path))

	disposition := `attachment;filename="` + ascii + `"`
	if ascii != utf {
		disposition += `;filename*=UTF-8''` + url.PathEscape(utf)
	}

	headers.Set("Content-Disposition", disposition)
	headers.Set("Content-Type", mime.TypeByExtension(ext))
}

func toURLPath(path string) string {
	if strings.HasPrefix(path, `/`) {
		return path[1:]
	}
	if strings.HasPrefix(path, `\\`) {
		return `\\` + filepath.ToSlash(path[2:])
	}
	return filepath.ToSlash(path)
}

func fromURLPath(path string) string {
	if filepath.Separator != '/' {
		return filepath.FromSlash(strings.TrimPrefix(path, "/"))
	}
	return path
}
