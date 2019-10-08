// +build !memfs

package main

import (
	"html/template"
	"net/http"
)

//go:generate memfsgen -tag memfs -pkg main assets assets_gen.go

var assetHandler = http.FileServer(http.Dir("assets"))

func assetTemplates() *template.Template {
	return template.Must(template.ParseGlob("assets/*.gohtml"))
}
