package main

import (
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

func uploadHandler(w http.ResponseWriter, r *http.Request) httpResult {
	if isLocalhost(r) {
		return httpResult{Status: http.StatusForbidden}
	}
	if err := r.ParseMultipartForm(100 << 20); err != nil {
		return httpResult{Status: http.StatusBadRequest, Error: err}
	}

	multipartFile := func(key string) *multipart.FileHeader {
		if r.MultipartForm == nil {
			return nil
		}
		vs := r.MultipartForm.File[key]
		if len(vs) == 0 {
			return nil
		}
		return vs[0]
	}
	multipartValue := func(key string) string {
		if r.MultipartForm == nil {
			return ""
		}
		vs := r.MultipartForm.Value[key]
		if len(vs) == 0 {
			return ""
		}
		return vs[0]
	}

	prefix := getPathPrefix(r)
	root := fromURLPath(multipartValue("root"), prefix)
	path := filepath.Join(root, filepath.FromSlash(multipartValue("path")))
	file, err := multipartFile("file").Open()
	if err != nil {
		return httpResult{Error: err}
	}
	defer file.Close()

	if fi, err := os.Stat(root); err != nil {
		return httpResult{Error: err}
	} else if !fi.IsDir() {
		return httpResult{Status: http.StatusForbidden}
	}

	if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
		return httpResult{Error: err}
	}

	dest, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		return httpResult{Error: err}
	}
	defer dest.Close()

	_, err = io.Copy(dest, file)
	if err != nil {
		return httpResult{Error: err}
	}

	err = dest.Close()
	if err != nil {
		return httpResult{Error: err}
	}
	return httpResult{Status: http.StatusOK}
}
