package main

import (
	"context"
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
)

var chrome = `C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`
var baseDir, dataDir, tempDir string
var lock *os.File

func main() {
	if exe, err := os.Executable(); err != nil {
		log.Fatal(err)
	} else {
		baseDir = filepath.Dir(exe)
		dataDir = filepath.Join(baseDir, "data")
		tempDir = filepath.Join(dataDir, "temp")
	}

	if err := os.MkdirAll(dataDir, 0700); err != nil {
		log.Fatal(err)
	}

	url := url.URL{Scheme: "http"}
	hideConsole()

	// path
	var path string
	if len(os.Args) > 1 {
		path = os.Args[1]
	} else {
		var err error
		path, err = os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
	}

	if fi, err := os.Stat(path); err != nil {
		log.Fatal(err)
	} else if abs, err := filepath.Abs(path); err != nil {
		log.Fatal(err)
	} else {
		if fi.IsDir() {
			url.Path = "/gallery/" + abs
		} else {
			url.Path = "/photo/" + abs
		}
	}

	if err := os.Chdir(baseDir); err != nil {
		log.Fatal(err)
	}

	// address
	ln, err := net.Listen("tcp", "[::1]:0")
	if err != nil {
		log.Fatal(err)
	}

	url.Host = ln.Addr().String()
	lock, err = createLock(url.Host)
	if err != nil {
		url.Host, err = getLocked()
		if err != nil {
			log.Fatal(err)
		}
		ln.Close()
	} else {
		exif := setupExifTool()
		http := setupHTTP()
		defer func() {
			log.Println("Exiting...")
			exif.Stop()
			http.Shutdown(context.Background())
			os.RemoveAll(tempDir)
		}()
		go http.Serve(ln)
	}

	chrome := setupChrome(url.String())
	if err := chrome.Run(); err != nil {
		log.Fatal(err)
	}
}

func setupChrome(url string) *exec.Cmd {
	dir := filepath.Join(dataDir, "chrome")

	prefs := filepath.Join(dir, "Default", "Preferences")
	if _, err := os.Stat(prefs); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(prefs), 0700); err == nil {
			ioutil.WriteFile(prefs, []byte(`{"download":{"prompt_for_download":true}}`), 0600)
		}
	}

	return exec.Command(chrome, "--app="+url, "--user-data-dir="+dir, "--no-first-run", "--disable-default-apps", "--disable-sync", "--disable-extensions", "--disable-plugins", "--disable-background-networking")
}

func createLock(address string) (file *os.File, err error) {
	filename := filepath.Join(dataDir, "lockfile")

	err = os.RemoveAll(filename)
	if err != nil {
		return
	}

	file, err = os.OpenFile(filename, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return
	}

	_, err = file.WriteString(address)
	return
}

func getLocked() (string, error) {
	filename := filepath.Join(dataDir, "lockfile")

	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}

	return string(buf), nil
}
