// +build !memfs

package main

import (
	"html/template"
	"net/http"
)

//go:generate go run github.com/ncruces/go-fs/memfsgen -minify -mimetype gohtml:text/html -tag memfs -pkg main assets assets_gen.go

var assets = http.Dir("assets")
var assetHandler = http.FileServer(assets)

func assetTemplates() *template.Template {
	return template.Must(template.ParseGlob("assets/*.gohtml"))
}
