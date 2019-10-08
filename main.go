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

	if err := os.Chdir(baseDir); err != nil {
		log.Fatal(err)
	}

	chrome := findChrome()
	if chrome != "" {
		hideConsole()
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

	sigs := make(chan os.Signal, 1)
	handleConsoleCtrl(sigs)
	signal.Notify(sigs)

	if chrome != "" {
		cmd := setupChrome(chrome, url.String())
		if err := cmd.Start(); err != nil {
			log.Fatal(err)
		}
		go func() {
			for {
				<-sigs
				exitChrome(cmd)
			}
		}()
		if err := cmd.Wait(); err != nil {
			log.Fatal(err)
		}
	} else {
		if err := openURLCmd(url.String()).Run(); err != nil {
			log.Fatal(err)
		}
		<-sigs
	}
}

func setupChrome(chrome, url string) *exec.Cmd {
	data := filepath.Join(dataDir, "chrome")
	cache := filepath.Join(tempDir, "chrome")

	prefs := filepath.Join(data, "Default", "Preferences")
	if _, err := os.Stat(prefs); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(prefs), 0755); err == nil {
			ioutil.WriteFile(prefs, []byte(`{"download":{"prompt_for_download":true}}`), 0666)
		}
	}

	return exec.Command(chrome, "--app="+url, "--user-data-dir="+data, "--disk-cache-dir="+cache, "--no-first-run",
		"--disable-default-apps", "--disable-sync", "--disable-extensions", "--disable-plugins",
		"--disable-bundled-ppapi-flash", "--disable-background-networking")
}
