package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ncruces/rethinkraw/internal/config"
)

var templates *template.Template

func setupHTTP() *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/gallery/", http.StripPrefix("/gallery", httpHandler(galleryHandler)))
	mux.Handle("/photo/", http.StripPrefix("/photo", httpHandler(photoHandler)))
	mux.Handle("/batch/", http.StripPrefix("/batch", httpHandler(batchHandler)))
	mux.Handle("/thumb/", http.StripPrefix("/thumb", httpHandler(thumbHandler)))
	mux.Handle("/dialog", httpHandler(dialogHandler))
	mux.Handle("/upload", httpHandler(uploadHandler))
	mux.Handle("/", assetHandler)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isLocalhost(r) {
			if !config.ServerMode || !matchHostServerName(r) {
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}
			if url := canUseTLS(r); url != "" {
				http.Redirect(w, r, url, http.StatusTemporaryRedirect)
				return
			}
			if _, pwd, _ := r.BasicAuth(); pwd != serverAuth {
				w.Header().Set("WWW-Authenticate", `Basic charset="UTF-8"`)
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			if r.URL.Path == "/" {
				http.Redirect(w, r, "/gallery", http.StatusTemporaryRedirect)
				return
			}
		}
		mux.ServeHTTP(w, r)
	})

	templates = assetTemplates()

	server := &http.Server{
		ReadHeaderTimeout: time.Second,
		IdleTimeout:       time.Minute,
		Handler:           handler,
	}
	return server
}

// httpResult helps httpHandler short circuit a result
type httpResult struct {
	Status   int
	Message  string
	Location string
	Error    error
}

func (r *httpResult) Done() bool { return r.Status != 0 || r.Location != "" || r.Error != nil }

// httpHandler is an http.Handler that returns an httpResult
type httpHandler func(w http.ResponseWriter, r *http.Request) httpResult

func (h httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch res := h(w, r); {

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
		templates.ExecuteTemplate(w, "error.gohtml", map[string]string{
			"Status":  http.StatusText(status),
			"Message": message,
		})
	} else {
		h.Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(message)
	}
}

func sendCached(w http.ResponseWriter, r *http.Request, path string) httpResult {
	if fi, err := os.Stat(path); err != nil {
		return httpResult{Error: err}
	} else {
		headers := w.Header()
		headers.Set("Last-Modified", fi.ModTime().UTC().Format(http.TimeFormat))
		if ims := r.Header.Get("If-Modified-Since"); ims != "" {
			if t, err := http.ParseTime(ims); err == nil {
				if fi.ModTime().Before(t.Add(time.Second)) {
					for k := range headers {
						switch k {
						case "Cache-Control", "Last-Modified":
							// keep
						default:
							delete(headers, k)
						}
					}
					return httpResult{Status: http.StatusNotModified}
				}
			}
		}
	}
	return httpResult{}
}

func sendAllowed(w http.ResponseWriter, r *http.Request, allowed ...string) httpResult {
	for _, method := range allowed {
		if method == r.Method {
			return httpResult{}
		}
	}

	w.Header().Set("Allow", strings.Join(allowed, ", "))
	return httpResult{Status: http.StatusMethodNotAllowed}
}

func isLocalhost(r *http.Request) bool {
	return r.TLS == nil && strings.TrimSuffix(r.Host, serverPort) == "localhost"
}

func getPathPrefix(r *http.Request) string {
	if !isLocalhost(r) {
		return serverPrefix
	}
	return ""
}

func toURLPath(path, prefix string) string {
	path = filepath.Clean(path)
	if strings.HasPrefix(path, prefix) {
		path = path[len(prefix):]
	} else {
		return ""
	}
	if filepath.Separator == '\\' && strings.HasPrefix(path, `\\`) {
		return `\\` + filepath.ToSlash(path[2:])
	}
	return strings.TrimPrefix(filepath.ToSlash(path), "/")
}

func fromURLPath(path, prefix string) string {
	if filepath.Separator != '/' {
		path = filepath.FromSlash(strings.TrimPrefix(path, "/"))
	}
	return filepath.Join(prefix, path)
}

func matchHostServerName(r *http.Request) bool {
	if r.TLS == nil {
		return true
	}
	host, _, err := net.SplitHostPort(r.Host)
	if err != nil {
		host = r.Host
	}
	return r.TLS.ServerName == host
}

func canUseTLS(r *http.Request) (url string) {
	if r.TLS != nil {
		return ""
	}

	host := r.Host
	name, port, err := net.SplitHostPort(host)
	if err != nil {
		return ""
	}
	if app := getAppDomain(name); app != "" {
		host = net.JoinHostPort(app, port)
		name = app
	}

	chi := tls.ClientHelloInfo{ServerName: name, SupportedVersions: []uint16{
		tls.VersionTLS13,
		tls.VersionTLS12,
	}}

	config := &serverConfig
	if config.GetConfigForClient != nil {
		cfg, err := config.GetConfigForClient(&chi)
		if err != nil {
			return ""
		}
		if cfg != nil {
			config = cfg
		}
	}

	if config.GetCertificate != nil {
		cert, err := config.GetCertificate(&chi)
		if err != nil {
			return ""
		}
		if cert != nil {
			u := r.URL
			u.Host = host
			u.Scheme = "https"
			return u.String()
		}
	}

	for _, cert := range config.Certificates {
		if chi.SupportsCertificate(&cert) == nil {
			u := r.URL
			u.Host = host
			u.Scheme = "https"
			return u.String()
		}
	}
	return ""
}

func getAppDomain(name string) string {
	if ip := net.ParseIP(name); ip != nil {
		name = ip.String()
		if ip4 := ip.To4(); len(ip4) == net.IPv4len {
			name = strings.ReplaceAll(name, ".", "-")
			return name + ".app.rethinkraw.com"

		} else if len(ip) == net.IPv6len {
			if strings.HasPrefix(name, "::") {
				name = "0" + name
			}
			if strings.HasSuffix(name, "::") {
				name = name + "0"
			}
			name = strings.ReplaceAll(name, ":", "-")
			return name + ".app.rethinkraw.com"
		}
	}
	return ""
}
