package exiftool

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"sync"
)

const boundary = "1854673209"

// Server wraps an instance of ExifTool that can process multiple commands sequentially.
// Servers avoid the overhead of loading ExifTool for each command.
// Servers are safe for concurrent use by multiple goroutines.
type Server struct {
	path   string
	args   []string
	srvMtx sync.Mutex
	cmdMtx sync.Mutex
	done   bool
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Scanner
	stderr *bufio.Scanner
}

// NewServer loads a new instance of ExifTool.
func NewServer(path, arg1 string, commonArg ...string) (*Server, error) {
	e := &Server{path: path}

	if arg1 != "" {
		e.args = append(e.args, arg1)
	}

	e.args = append(e.args, "-stay_open", "true", "-@", "-", "-common_args", "-echo4", "{ready"+boundary+"}", "-charset", "filename=utf8")
	e.args = append(e.args, commonArg...)

	return e, e.start()
}

func (e *Server) start() error {
	cmd := exec.Command(e.path, e.args...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	e.stdin = stdin
	e.stdout = bufio.NewScanner(stdout)
	e.stderr = bufio.NewScanner(stderr)
	e.stdout.Split(splitReadyToken)
	e.stderr.Split(splitReadyToken)

	err = cmd.Start()
	if err != nil {
		return err
	}

	e.cmd = cmd
	return nil
}

func (e *Server) restart() {
	e.srvMtx.Lock()
	defer e.srvMtx.Unlock()
	if e.done {
		return
	}

	e.cmd.Process.Kill()
	e.cmd.Process.Release()
	e.start()
}

// Close causes ExifTool to exit immediately.
// Close does not wait until ExifTool has actually exited.
func (e *Server) Close() error {
	e.srvMtx.Lock()
	defer e.srvMtx.Unlock()
	if e.done {
		return nil
	}

	err := e.cmd.Process.Kill()
	e.cmd.Process.Release()
	e.done = true
	return err
}

// Shutdown gracefully shuts down ExifTool without interrupting any commands,
// and waits for it to complete.
func (e *Server) Shutdown() error {
	e.cmdMtx.Lock()
	defer e.cmdMtx.Unlock()

	fmt.Fprintln(e.stdin, "-stay_open")
	fmt.Fprintln(e.stdin, "false")
	e.stdin.Close()

	err := e.cmd.Wait()
	return err
}

// Command runs an ExifTool command with the given arguments and returns its stdout.
// Commands should neither read from stdin, nor write binary data to stdout.
func (e *Server) Command(arg ...string) ([]byte, error) {
	e.cmdMtx.Lock()
	defer e.cmdMtx.Unlock()

	for _, a := range arg {
		fmt.Fprintln(e.stdin, a)
	}

	_, err := fmt.Fprintln(e.stdin, "-execute"+boundary)
	if err != nil {
		e.restart()
		return nil, err
	}

	if !e.stdout.Scan() {
		err := e.stdout.Err()
		if err == nil {
			err = io.EOF
		}
		e.restart()
		return nil, err
	}
	if !e.stderr.Scan() {
		err := e.stderr.Err()
		if err == nil {
			err = io.EOF
		}
		e.restart()
		return nil, err
	}

	if len(e.stderr.Bytes()) > 0 {
		return nil, fmt.Errorf("exiftool: %s", bytes.TrimSpace(e.stderr.Bytes()))
	}
	res := make([]byte, len(e.stdout.Bytes()))
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
