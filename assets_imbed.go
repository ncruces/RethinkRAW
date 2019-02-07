// +build imbed

package main

import (
	"html/template"
	"net/http"

	"./imbed"
)

var assetHandler = http.HandlerFunc(imbed.ServeHTTP)

func assetTemplates() *template.Template {
	t := template.New("")

	for _, f := range []string{"gallery.gohtml", "photo.gohtml", "error.gohtml"} {
		template.Must(t.New(f).Parse(imbed.Must(f).String()))
	}

	return t
}
