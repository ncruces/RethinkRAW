package exiftool

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"syscall"

	"errors"
)

const boundary = "1854673209"

type Server struct {
	path   string
	args   []string
	mtx    sync.Mutex
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Scanner
	stderr *bufio.Scanner
}

func NewServer(path, arg1 string, commonArg ...string) (*Server, error) {
	e := &Server{path: path}

	if arg1 != "" {
		e.args = append(e.args, arg1)
	}

	e.args = append(e.args, "-stay_open", "true", "-@", "-", "-common_args", "-echo4", "{ready"+boundary+"}", "-charset", "filename=utf8")
	e.args = append(e.args, commonArg...)

	return e, e.start()
}

func (e *Server) start() (err error) {
	cmd := exec.Command(e.path, e.args...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return
	}

	e.stdin = stdin
	e.stdout = bufio.NewScanner(stdout)
	e.stderr = bufio.NewScanner(stderr)
	e.stdout.Split(splitReadyToken)
	e.stderr.Split(splitReadyToken)

	err = cmd.Start()
	if err != nil {
		return
	}

	e.cmd = cmd
	return nil
}

func (e *Server) restart() {
	e.cmd.Process.Signal(syscall.SIGTERM)
	e.stdin.Close()
	e.cmd = nil
	e.start()
}

func (e *Server) Shutdown() (err error) {
	e.mtx.Lock()
	defer e.mtx.Unlock()

	if e.cmd != nil {
		e.cmd = nil
		fmt.Fprintln(e.stdin, "-stay_open")
		fmt.Fprintln(e.stdin, "false")
		err = e.stdin.Close()
	}

	return
}

func (e *Server) Command(arg ...string) (res []byte, err error) {
	e.mtx.Lock()
	defer e.mtx.Unlock()

	if e.cmd == nil {
		return nil, errors.New("server closed")
	}

	for _, a := range arg {
		fmt.Fprintln(e.stdin, a)
	}

	_, err = fmt.Fprintln(e.stdin, "-execute"+boundary)
	if err != nil {
		e.restart()
		return
	}

	if !e.stdout.Scan() {
		err = e.stdout.Err()
		if err == nil {
			err = io.EOF
		}
		e.restart()
		return
	}
	if !e.stderr.Scan() {
		err = e.stderr.Err()
		if err == nil {
			err = io.EOF
		}
		e.restart()
		return
	}

	if len(e.stderr.Bytes()) > 0 {
		return nil, errors.New(string(bytes.TrimSpace(e.stderr.Bytes())))
	}
	res = make([]byte, len(e.stdout.Bytes()))
	copy(res, e.stdout.Bytes())
	return res, nil
}

func splitReadyToken(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if i := bytes.Index(data, []byte("{ready"+boundary+"}")); i >= 0 {
		if n := bytes.IndexByte(data[i:], '\n'); n >= 0 {
			if atEOF && len(data) == (n+i+1) { // nothing left to scan
				return n + i + 1, data[:i], bufio.ErrFinalToken
			} else {
				return n + i + 1, data[:i], nil
			}
		}
	}

	if atEOF {
		return 0, data, io.EOF
	}

	return 0, nil, nil
}
