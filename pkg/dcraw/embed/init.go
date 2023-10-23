// Package embed provides support embeding dcraw into your application.
//
// You can obtain this build of dcraw from:
// https://github.com/ncruces/dcraw/releases/tag/v9.28.6-wasm
//
// Before importing this package, inspect the dcraw license and
// consider the implications.
package embed

import (
	_ "embed"

	"github.com/ncruces/rethinkraw/pkg/dcraw"
)

//go:generate -command go-fetch go run github.com/ncruces/go-fetch
//go:generate go-fetch -unpack "https://github.com/ncruces/dcraw/releases/download/v9.28.8-wasm/dcraw.wasm.gz" dcraw.wasm

//go:embed dcraw.wasm
var binary []byte

func init() {
	dcraw.Binary = binary
}
