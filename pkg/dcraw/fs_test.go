package dcraw

import (
	"strings"
	"testing"
	"testing/fstest"
)

func Test_readerFS(t *testing.T) {
	fs := readerFS{strings.NewReader("contents")}
	if err := fstest.TestFS(fs, readerFSname); err != nil {
		t.Fatal(err)
	}
}
