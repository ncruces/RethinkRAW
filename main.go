package main

import (
	"context"
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
)

var baseDir, dataDir, tempDir string
var lock *os.File

func main() {
	if err := setupDirs(); err != nil {
		log.Fatal(err)
	}

	if err := loadConfig(); err != nil {
		log.Fatal(err)
	}

	url := url.URL{Scheme: "http"}

	if len(os.Args) > 1 {
		if fi, err := os.Stat(os.Args[1]); err != nil {
			log.Fatal(err)
		} else if abs, err := filepath.Abs(os.Args[1]); err != nil {
			log.Fatal(err)
		} else {
			if fi.IsDir() {
				url.Path = "/gallery/" + toURLPath(abs)
			} else {
				url.Path = "/photo/" + toURLPath(abs)
			}
		}
	}

	if err := testDNGConverter(); err != nil {
		url.Path = "/dngconv.html"
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
			http.Shutdown(context.Background())
			exif.Stop()
			os.RemoveAll(tempDir)
		}()
		go http.Serve(ln)
	}

	if chrome := setupChrome(url.String()); chrome != nil {
		hideConsole()
		if err := chrome.Run(); err != nil {
			log.Fatal(err)
		}
	} else {
		if err := openURLCmd(url.String()).Run(); err != nil {
			log.Fatal(err)
		}
		c := make(chan os.Signal, 1)
		signal.Notify(c)
		<-c
	}
}

func setupDirs() error {
	if exe, err := os.Executable(); err != nil {
		return err
	} else {
		baseDir = filepath.Dir(exe)
		tempDir = filepath.Join(os.TempDir(), "RethinkRAW")
	}

	if err := os.Chdir(baseDir); err != nil {
		return err
	}

	dataDir = filepath.Join(baseDir, "data")
	if err := testDataDir(dataDir); err == nil {
		return err
	}

	dataDir = filepath.Join(os.Getenv("APPDATA"), "RethinkRAW")
	return testDataDir(dataDir)
}

func testDataDir(dir string) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	if f, err := os.Create(filepath.Join(dir, "lastrun")); err != nil {
		return err
	} else {
		return f.Close()
	}
}

func setupChrome(url string) *exec.Cmd {
	chrome := getChromePath()
	if chrome == "" {
		return nil
	}

	data := filepath.Join(dataDir, "chrome")
	cache := filepath.Join(tempDir, "chrome")

	prefs := filepath.Join(data, "Default", "Preferences")
	if _, err := os.Stat(prefs); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(prefs), 0700); err == nil {
			ioutil.WriteFile(prefs, []byte(`{"download":{"prompt_for_download":true}}`), 0600)
		}
	}

	return exec.Command(chrome, "--app="+url, "--user-data-dir="+data, "--disk-cache-dir="+cache, "--no-first-run",
		"--disable-default-apps", "--disable-sync", "--disable-extensions", "--disable-plugins", "--disable-translate",
		"--disable-background-networking")
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
