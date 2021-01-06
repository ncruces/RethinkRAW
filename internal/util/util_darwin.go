package util

import (
	"os"
	"strings"
)

func init() {
	// ignore Process Serial Number argument
	for i, a := range os.Args {
		if strings.HasPrefix(a, "-psn_") {
			os.Args = append(os.Args[:i], os.Args[i+1:]...)
			break
		}
	}
}

func HideConsole() {}
