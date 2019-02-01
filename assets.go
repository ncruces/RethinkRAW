// +build !vfsdata

package main

//go:generate go run assets_vfsgen.go

import "net/http"

var assets = http.Dir("assets")
