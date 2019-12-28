package main

import (
	"net/http"
	"os"

	"github.com/ncruces/go-ui/dialog"
)

func configHandler(w http.ResponseWriter, r *http.Request) HTTPResult {
	if r.ParseForm() != nil {
		return HTTPResult{Status: http.StatusBadRequest}
	}

	_, dngconv := r.Form["dngconv"]
	if dngconv {
		bringToTop()
		if file, err := dialog.OpenFile("", os.Getenv("PROGRAMFILES"), []dialog.FileFilter{{Exts: []string{".exe"}}}); err != nil {
			return HTTPResult{Error: err}
		} else if file == "" {
			return HTTPResult{Status: http.StatusResetContent}
		} else if err := testDNGConverter(file); err != nil {
			return HTTPResult{Error: err}
		} else {
			serverConfig.DNGConverter = file
			if err := saveConfig(); err != nil {
				return HTTPResult{Error: err}
			}
			return HTTPResult{Location: "/"}
		}
	}

	return HTTPResult{Status: http.StatusInternalServerError}
}
