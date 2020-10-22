package exiftool

import (
	"bytes"
	"io/ioutil"
	"strconv"
	"testing"
)

func TestCommandAsync(t *testing.T) {
	in, out, err := CommandAsync(path, arg1, "-@", "-")
	if err != nil {
		t.Fatal(err)
	}

	// ask for version number
	_, err = in.Write([]byte("-ver\n"))
	if err != nil {
		t.Fatal(err)
	}
	err = in.Close()
	if err != nil {
		t.Error(err)
	}

	dat, err := ioutil.ReadAll(out)
	if err != nil {
		t.Fatal(err)
	}
	err = out.Close()
	if err != nil {
		t.Error(err)
	}

	if ver, err := strconv.ParseFloat(string(bytes.TrimSpace(dat)), 64); err != nil {
		t.Error(err)
	} else {
		t.Log(ver)
	}
}
