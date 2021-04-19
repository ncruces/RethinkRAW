// +build memfs

package main

import (
	"html/template"
)

var assetHandler = assets

func assetTemplates() *template.Template {
	return template.Must(template.ParseFS(assets, "*.gohtml"))
}
