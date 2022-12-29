// Package embed provides support embeding dcraw into your application.
//
// You can obtain this build of dcraw from:
// https://github.com/ncruces/dcraw/releases/tag/v9.28.5-wasm
//
// Before importing this package, inspect the dcraw license and
// consider the implications.
package embed

import (
	_ "embed"

	"github.com/ncruces/rethinkraw/pkg/dcraw"
)

//go:embed dcraw.wasm
var binary []byte

func init() {
	dcraw.Binary = binary
}
