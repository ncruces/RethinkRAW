package main

import (
	"net/http"
	"os"

	nfd "github.com/ncruces/go-nativefiledialog"
)

func configHandler(w http.ResponseWriter, r *http.Request) HTTPResult {
	r.ParseForm()

	_, dngconv := r.Form["dngconv"]
	if dngconv {
		bringToTop()
		if file, err := nfd.OpenDialog("exe", os.Getenv("PROGRAMFILES")); err != nil {
			return handleError(err)
		} else if file == "" {
			return HTTPResult{Status: http.StatusResetContent}
		} else if err := testDNGConverter(file); err != nil {
			return handleError(err)
		} else {
			serverConfig.DNGConverter = file
			saveConfig()
			return HTTPResult{Location: "/"}
		}
	}

	return HTTPResult{Status: http.StatusInternalServerError}
}
