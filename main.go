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
	"syscall"

	"rethinkraw/osutil"
)

var shutdown = make(chan os.Signal, 1)

func init() {
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	hideConsole()
}

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	if err := setupPaths(); err != nil {
		return err
	}

	url := url.URL{
		Scheme: "http",
		Host:   "[::1]:39639",
	}

	if len(os.Args) > 1 {
		if fi, err := os.Stat(os.Args[1]); err != nil {
			return err
		} else if abs, err := filepath.Abs(os.Args[1]); err != nil {
			return err
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

	var server bool
	if ln, err := net.Listen("tcp", url.Host); err == nil {
		server = true
		http := setupHTTP()
		exif, err := setupExifTool()
		if err != nil {
			return err
		}
		defer func() {
			http.Shutdown(context.Background())
			exif.Shutdown()
			os.RemoveAll(tempDir)
		}()
		go http.Serve(ln)
	}

	if chrome := findChrome(); chrome != "" {
		cmd := setupChrome(chrome, url.String())
		if err := cmd.Start(); err != nil {
			return err
		}

		go func() {
			<-shutdown
			exitChrome(cmd)
		}()
		if err := cmd.Wait(); err != nil {
			return err
		}
	} else {
		if err := osutil.ShellOpen(url.String()); err != nil {
			return err
		}
		if server {
			<-shutdown
		}
	}

	return nil
}

func setupChrome(chrome, url string) *exec.Cmd {
	data := filepath.Join(dataDir, "chrome")
	cache := filepath.Join(tempDir, "chrome")

	prefs := filepath.Join(data, "Default", "Preferences")
	if _, err := os.Stat(prefs); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(prefs), 0755); err == nil {
			ioutil.WriteFile(prefs, []byte(`{
				"profile": {"block_third_party_cookies": true},
				"download": {"prompt_for_download": true},
				"enable_do_not_track": true
			}`), 0666)
		}
	}

	// https://source.chromium.org/chromium/chromium/src/+/master:chrome/test/chromedriver/chrome_launcher.cc
	return exec.Command(chrome, "--app="+url, "--homepage="+url, "--user-data-dir="+data, "--disk-cache-dir="+cache,
		"--no-first-run", "--no-service-autorun", "--disable-sync", "--disable-extensions", "--disable-default-apps",
		"--disable-background-networking", "--disable-client-side-phishing-detection")
}
