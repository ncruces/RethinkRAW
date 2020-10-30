// +build !windows
// +build !darwin

package osutil

import "os"

func isHidden(fi os.FileInfo) bool {
	return false
}
