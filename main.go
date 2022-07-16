package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/ncruces/rethinkraw/internal/config"
	"github.com/ncruces/rethinkraw/internal/optls"
	"github.com/ncruces/rethinkraw/pkg/chrome"
	"github.com/ncruces/rethinkraw/pkg/osutil"
	"github.com/ncruces/zenity"
)

var shutdown = make(chan os.Signal, 1)

var (
	serverHost   string
	serverPort   string
	serverPrefix string
	serverConfig tls.Config
)

func init() {
	signal.Notify(shutdown, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	osutil.CreateConsole()
	osutil.CleanupArgs()
	log.SetOutput(os.Stderr)
}

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	if err := config.SetupPaths(); err != nil {
		return err
	}

	port := flag.Int("port", 39639, "the port on which the server listens for connections")
	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "usage: %s [OPTION]... DIRECTORY\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()

	serverPort = ":" + strconv.Itoa(*port)
	var url url.URL

	if config.ServerMode {
		if flag.NArg() != 1 {
			flag.Usage()
			os.Exit(2)
		}
		if fi, err := os.Stat(flag.Arg(0)); err != nil {
			return err
		} else if abs, err := filepath.Abs(flag.Arg(0)); err != nil {
			return err
		} else if fi.IsDir() {
			serverPrefix = abs
		} else {
			flag.Usage()
			os.Exit(2)
		}

		if err := testDNGConverter(); err != nil {
			return err
		}
	} else {
		serverHost = "localhost"
		url.Scheme = "http"
		url.Host = serverHost + serverPort

		if flag.NArg() > 0 {
			if fi, err := os.Stat(flag.Arg(0)); err != nil {
				return err
			} else if abs, err := filepath.Abs(flag.Arg(0)); err != nil {
				return err
			} else if flag.NArg() > 1 {
				url.Path = "/batch/" + toBatchPath(flag.Args()...)
			} else if fi.IsDir() {
				url.Path = "/gallery/" + toURLPath(abs, "")
			} else {
				url.Path = "/photo/" + toURLPath(abs, "")
			}
		}

		if err := testDNGConverter(); err != nil {
			url.Path = "/dngconv.html"
		}
	}

	if ln, err := optls.Listen("tcp", serverHost+serverPort, &serverConfig); err == nil {
		http := setupHTTP()
		exif, err := setupExifTool()
		if err != nil {
			return err
		}
		defer func() {
			http.Shutdown(context.Background())
			exif.Shutdown()
			os.RemoveAll(config.TempDir)
		}()
		go http.Serve(ln)
	} else if config.ServerMode {
		return err
	}

	if config.ServerMode {
		<-shutdown
	} else if chrome.IsInstalled() {
		data := filepath.Join(config.DataDir, "chrome")
		cache := filepath.Join(config.TempDir, "chrome")
		cmd := chrome.Command(url.String(), data, cache)

		if err := cmd.Start(); err != nil {
			return err
		}
		go func() {
			for s := range shutdown {
				cmd.Signal(s)
			}
		}()
		return cmd.Wait()
	} else {
		return zenity.Error(
			"Please download and install either Google Chrome or Microsoft Edge.",
			zenity.Title("Google Chrome not found"))
	}
	return nil
}
