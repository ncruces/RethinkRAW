package exiftool

import (
	"bytes"
	"errors"
	"io"
	"os/exec"
)

// Command runs an ExifTool command with the given arguments asynchronously,
// and returns pipes connected to its stdin and stdout.
// Writing to stdin and reading from stdout should be done concurrently.
func CommandAsync(path, arg1 string, arg ...string) (stdin io.WriteCloser, stdout io.ReadCloser, err error) {
	var args []string

	if arg1 != "" {
		args = append(args, arg1)
	}

	args = append(args, "-charset", "filename=utf8")
	args = append(args, arg...)

	var res asyncResult

	res.cmd = exec.Command(path, args...)
	res.cmd.Stderr = &res.err
	res.out, err = res.cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}
	stdin, err = res.cmd.StdinPipe()
	if err != nil {
		return nil, nil, err
	}

	err = res.cmd.Start()
	if err != nil {
		return nil, nil, err
	}
	return stdin, &res, nil
}

type asyncResult struct {
	cmd *exec.Cmd
	out io.ReadCloser
	err bytes.Buffer
}

func (res *asyncResult) Read(p []byte) (int, error) {
	n, err := res.out.Read(p)
	if err == io.EOF {
		cerr := res.Close()
		if cerr != nil {
			return n, cerr
		}
	}
	return n, err
}

func (res *asyncResult) Close() error {
	if res.cmd != nil {
		err := res.cmd.Wait()
		var eerr *exec.ExitError
		if errors.As(err, &eerr) {
			eerr.Stderr = res.err.Bytes()
		}
		res.cmd = nil
		return err
	}
	return nil
}
