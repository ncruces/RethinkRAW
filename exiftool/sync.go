package exiftool

import (
	"io"
	"os/exec"
)

// Command runs an ExifTool command with the given arguments and stdin and returns its stdout.
func Command(path, arg1 string, stdin io.Reader, arg ...string) (stdout []byte, err error) {
	var args []string

	if arg1 != "" {
		args = append(args, arg1)
	}

	args = append(args, "-charset", "filename=utf8")
	args = append(args, arg...)

	cmd := exec.Command(path, args...)
	cmd.Stdin = stdin
	return cmd.Output()
}
