// +build !imbed

package main

import (
	"html/template"
	"net/http"
)

//go:generate go-imbed assets imbed

var assetHandler = http.FileServer(http.Dir("assets"))

func assetTemplates() *template.Template {
	return template.Must(template.ParseGlob("assets/*.gohtml"))
}
