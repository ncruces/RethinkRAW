package nfd

import (
	"testing"
)

func TestOpenDialog(t *testing.T) {
	t.Log(OpenDialog("png,jpg;pdf", ""))
}

func TestOpenDialogMultiple(t *testing.T) {
	t.Log(OpenDialogMultiple("png,jpg;pdf", ""))
}

func TestSaveDialog(t *testing.T) {
	t.Log(SaveDialog("png,jpg;pdf", ""))
}

func TestPickFolder(t *testing.T) {
	t.Log(PickFolder(""))
}
