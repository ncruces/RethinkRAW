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

func Installed() bool {
	once.Do(findChrome)
	return chrome != ""
}

type Cmd struct {
	cmd *exec.Cmd
}

func New(url string, dataDir, tempDir string) *Cmd {
	once.Do(findChrome)
	if chrome == "" {
		return nil
	}

	prefs := filepath.Join(dataDir, "Default", "Preferences")
	if _, err := os.Stat(prefs); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(prefs), 0755); err == nil {
			ioutil.WriteFile(prefs, []byte(`{
				"profile": {"block_third_party_cookies": true},
				"enable_do_not_track": true
			}`), 0666)
		}
	}

	// https://source.chromium.org/chromium/chromium/src/+/master:chrome/test/chromedriver/chrome_launcher.cc
	cmd := exec.Command(chrome, "--app="+url, "--homepage="+url, "--user-data-dir="+dataDir, "--disk-cache-dir="+tempDir,
		"--no-first-run", "--no-service-autorun", "--disable-sync", "--disable-extensions", "--disable-default-apps",
		"--disable-background-networking", "--disable-client-side-phishing-detection")
	return &Cmd{cmd}
}

func (c *Cmd) Run() error {
	return c.cmd.Run()
}

func (c *Cmd) Start() error {
	return c.cmd.Start()
}

func (c *Cmd) Wait() error {
	return c.cmd.Wait()
}

func (c *Cmd) Pid() int {
	return c.cmd.Process.Pid
}
