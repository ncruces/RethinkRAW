package chrome

import (
	"os"
	"path/filepath"
)

func findChrome() {
	versions := []string{"Google Chrome", "Chromium"}

	for _, v := range versions {
		c := filepath.Join("/Applications", v+".app", "Contents/MacOS", v)
		if _, err := os.Stat(c); err == nil {
			chrome = c
			return
		}
	}
}

func exitProcess(p *os.Process) error {
	return p.Signal(os.Interrupt)
}