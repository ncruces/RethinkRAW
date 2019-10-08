// +build memfs

package main

import (
	"html/template"

	"github.com/shurcooL/httpfs/html/vfstemplate"
)

var assetHandler = assets

func assetTemplates() *template.Template {
	return template.Must(vfstemplate.ParseGlob(assets, nil, "/*.gohtml"))
}
