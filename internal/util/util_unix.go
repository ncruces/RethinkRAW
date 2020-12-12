// +build !windows
// +build !darwin

package util

func GetANSIPath(path string) (string, error) {
	return path, nil
}

func BringToTop() {}

func HideConsole() {}
