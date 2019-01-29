package exiftool

import (
	"io"
	"os/exec"
	"sync"
)

type AsyncResult struct {
	wg  sync.WaitGroup
	Out []byte
	Err error
}

func (r *AsyncResult) Get() ([]byte, error) {
	r.wg.Wait()
	return r.Out, r.Err
}

func CommandAsync(path string, r io.Reader, arg ...string) *AsyncResult {
	var args []string

	args = append(args, "-charset", "filename=utf8")
	args = append(args, arg...)

	cmd := exec.Command(path, args...)
	cmd.Stdin = r

	var res AsyncResult
	res.wg.Add(1)
	go func() {
		res.Out, res.Err = cmd.Output()
		res.wg.Done()
	}()
	return &res
}
