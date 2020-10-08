package main

import (
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/ncruces/zenity"
)

func configHandler(w http.ResponseWriter, r *http.Request) HTTPResult {
	if r.ParseForm() != nil {
		return HTTPResult{Status: http.StatusBadRequest}
	}

	_, dngconv := r.Form["dngconv"]
	if dngconv {
		bringToTop()
		var opts []zenity.Option
		switch runtime.GOOS {
		case "windows":
			opts = append(opts, zenity.Filename(os.Getenv("PROGRAMFILES")), zenity.FileFilters{
				{Name: "Applications", Patterns: []string{"*.exe"}},
			})
		case "darwin":
			opts = append(opts, zenity.Filename("/Applications"), zenity.FileFilters{
				{Name: "Applications", Patterns: []string{"*.app"}},
			})
		}
		if file, err := zenity.SelectFile(opts...); err != nil {
			return HTTPResult{Error: err}
		} else if file == "" {
			return HTTPResult{Status: http.StatusResetContent}
		} else if err := testDNGConverter(file); err != nil {
			return HTTPResult{Error: err}
		} else {
			if runtime.GOOS == "darwin" && strings.HasSuffix(file, ".app") {
				file += "/Contents/MacOS/Adobe DNG Converter"
			}
			serverConfig.DNGConverter = file
			if err := saveConfig(); err != nil {
				return HTTPResult{Error: err}
			}
			return HTTPResult{Location: "/"}
		}
	}

	return HTTPResult{Status: http.StatusInternalServerError}
}
