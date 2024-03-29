//go:build !memfs

package main

import (
	"html/template"
	"net/http"
	"os"
)

//go:generate -command memfsgen go run github.com/ncruces/go-fs/memfsgen
//go:generate memfsgen -minify -mimetype gohtml:text/html -tag memfs -pkg main assets assets_gen.go

var assets = os.DirFS("assets")
var assetHandler = http.FileServer(http.Dir("assets"))

func assetTemplates() *template.Template {
	return template.Must(template.ParseGlob("assets/*.gohtml"))
}
