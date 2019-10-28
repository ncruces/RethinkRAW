// +build !memfs

package main

import (
	"html/template"
	"net/http"
)

//go:generate memfsgen -tag memfs -pkg main -minify assets assets_gen.go

var assets = http.Dir("assets")
var assetHandler = http.FileServer(assets)

func assetTemplates() *template.Template {
	return template.Must(template.ParseGlob("assets/*.gohtml"))
}
