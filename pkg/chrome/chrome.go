// Package chrome provides support to locate and run Google Chrome.
package chrome

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

var once sync.Once
var chrome string

// IsInstalled checks if Chrome is installed.
func IsInstalled() bool {
	once.Do(findChrome)
	return chrome != ""
}

// Cmd represents a Chrome instance being prepared or run.
type Cmd exec.Cmd

// Command returns the Cmd struct to execute a Chrome app loaded from url,
// and with the given user data and disk cache directories.
func Command(url string, dataDir, cacheDir string) *Cmd {
	once.Do(findChrome)
	if chrome == "" {
		return nil
	}

	prefs := filepath.Join(dataDir, "Default", "Preferences")
	if _, err := os.Stat(prefs); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(prefs), 0700); err == nil {
			ioutil.WriteFile(prefs, []byte(`{
				"profile": {"cookie_controls_mode": 1},
				"search":  {"suggest_enabled": false},
				"signin":  {"allowed_on_next_startup": false},
				"enable_do_not_track": true
			}`), 0600)
		}
	}
	local := filepath.Join(dataDir, "Local State")
	if _, err := os.Stat(local); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(local), 0700); err == nil {
			ioutil.WriteFile(local, []byte(`{
				"profiles": {
					"edge_implicitly_signed_in": [{
						"edge_account_type": 1
					}]
				}
			}`), 0600)
		}
	}

	// https://source.chromium.org/chromium/chromium/src/+/master:chrome/test/chromedriver/chrome_launcher.cc
	cmd := exec.Command(chrome, "--app="+url, "--homepage="+url, "--user-data-dir="+dataDir, "--disk-cache-dir="+cacheDir,
		"--no-first-run", "--no-service-autorun", "--disable-sync", "--disable-extensions", "--disable-default-apps",
		"--disable-background-networking", "--disable-client-side-phishing-detection")
	return (*Cmd)(cmd)
}

// Run starts Chrome and waits for it to complete.
func (c *Cmd) Run() error {
	return (*exec.Cmd)(c).Run()
}

// Start starts Chrome but does not wait for it to complete.
func (c *Cmd) Start() error {
	return (*exec.Cmd)(c).Start()
}

// Wait for Chrome to exit.
func (c *Cmd) Wait() error {
	return (*exec.Cmd)(c).Wait()
}

// Signal sends a signal to Chrome.
func (c *Cmd) Signal(sig os.Signal) error {
	return signal(c.Process, sig)
}
