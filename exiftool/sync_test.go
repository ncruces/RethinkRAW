package exiftool

import (
	"bytes"
	"strconv"
	"testing"
)

func TestCommand(t *testing.T) {
	// ask for version number
	out, err := Command(path, arg1, nil, "-ver")
	if err != nil {
		t.Fatal(err)
	} else if ver, err := strconv.ParseFloat(string(bytes.TrimSpace(out)), 64); err != nil {
		t.Error(err)
	} else {
		t.Log(ver)
	}
}
