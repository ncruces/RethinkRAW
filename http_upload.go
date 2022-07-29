package main

import "net/http"

func uploadHandler(w http.ResponseWriter, r *http.Request) httpResult {
	if isLocalhost(r) {
		return httpResult{Status: http.StatusForbidden}
	}
	if err := r.ParseForm(); err != nil {
		return httpResult{Status: http.StatusBadRequest, Error: err}
	}
	return httpResult{Status: http.StatusNotFound}
}
