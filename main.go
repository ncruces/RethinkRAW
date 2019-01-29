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

func main() {
	if err := setupDirs(); err != nil {
		log.Fatal(err)
	}

	if err := loadConfig(); err != nil {
		log.Fatal(err)
	}

	url := url.URL{
		Scheme: "http",
		Host:   "[::1]:39639",
	}

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

	ln, err := net.Listen("tcp", url.Host)
	if err == nil {
		exif := setupExifTool()
		http := setupHTTP()
		defer func() {
			http.Shutdown(context.Background())
			exif.Shutdown()
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
		handleConsoleCtrl(c)
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
