// Package chrome provides support to locate and run Google Chrome.
package chrome

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/ncruces/jason"
	"golang.org/x/net/websocket"
)

var once sync.Once
var chrome string

// IsInstalled checks if Chrome is installed.
func IsInstalled() bool {
	once.Do(findChrome)
	return chrome != ""
}

// Cmd represents a Chrome instance being prepared or run.
type Cmd struct {
	cmd *exec.Cmd
	ws  *websocket.Conn
	msg uint32
}

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
						"edge_account_type": 1,
						"id": "0000000000000000"
					}]
				}
			}`), 0600)
		}
	}

	// https://github.com/GoogleChrome/chrome-launcher/blob/master/docs/chrome-flags-for-tools.md
	// https://source.chromium.org/chromium/chromium/src/+/master:chrome/test/chromedriver/chrome_launcher.cc
	cmd := exec.Command(chrome, "--app="+url, "--homepage="+url, "--user-data-dir="+dataDir, "--disk-cache-dir="+cacheDir,
		"--bwsi", "--no-first-run", "--no-default-browser-check", "--no-service-autorun",
		"--disable-sync", "--disable-breakpad", "--disable-extensions", "--disable-default-apps",
		"--disable-component-extensions-with-background-pages", "--disable-background-networking",
		"--disable-domain-reliability", "--disable-client-side-phishing-detection", "--disable-component-update")
	return &Cmd{cmd: cmd}
}

// Run starts Chrome and waits for it to complete.
func (c *Cmd) Run() error {
	if err := c.Start(); err != nil {
		return err
	}
	return c.Wait()
}

// Start starts Chrome but does not wait for it to complete.
func (c *Cmd) Start() error {
	c.cmd.Args = append(c.cmd.Args, "--remote-debugging-port=0")
	pipe, err := c.cmd.StderrPipe()
	if err != nil {
		return err
	}
	defer pipe.Close()

	err = c.cmd.Start()
	if err != nil {
		return err
	}

	var targets = set[string]{}
	scan := bufio.NewScanner(pipe)
	for scan.Scan() {
		const prefix = "DevTools listening on "
		line := scan.Bytes()
		if bytes.HasPrefix(line, []byte(prefix)) {
			url := line[len(prefix):]
			c.ws, err = websocket.Dial(string(url), "", "http://localhost")
			if err != nil {
				return err
			}
			c.send("Target.setDiscoverTargets", "", jason.Object{"discover": true})
			go func() {
				for {
					var msg cdpMessage
					err := websocket.JSON.Receive(c.ws, &msg)
					if err == io.EOF {
						break
					}
					if err != nil {
						log.Print(err)
					}
					switch msg.Method {
					case "Target.targetDestroyed", "Target.targetCrashed":
						targets.Del(jason.To[string](msg.Params["targetId"]))
						if len(targets) <= 1 {
							c.send("Browser.close", "", nil)
						}
					case "Target.targetCreated":
						info := jason.To[cdpTargetInfo](msg.Params["targetInfo"])
						targets.Add(info.TargetID)
						switch info.URL {
						case "about:blank", "about://newtab/", "chrome://newtab/", "edge://newtab/":
							c.send("Target.closeTarget", "", jason.Object{
								"targetId": info.TargetID,
							})
						}
					}
				}
			}()
			break
		}
	}
	return scan.Err()
}

// Wait for Chrome to exit.
func (c *Cmd) Wait() error {
	return c.cmd.Wait()
}

// Signal sends a signal to Chrome.
func (c *Cmd) Signal(sig os.Signal) error {
	return signal(c.cmd.Process, sig)
}

// Close closes Chrome.
func (c *Cmd) Close() error {
	return c.send("Browser.close", "", nil)
}

func (c *Cmd) send(method, session string, params any) error {
	return websocket.JSON.Send(c.ws, jason.Object{
		"id":        atomic.AddUint32(&c.msg, 1),
		"method":    method,
		"params":    params,
		"sessionId": session,
	})
}

type cdpMessage struct {
	ID     uint32          `json:"id,omitempty"`
	Method string          `json:"method,omitempty"`
	Result jason.Raw       `json:"result,omitempty"`
	Params jason.RawObject `json:"params,omitempty"`
}

type cdpTargetInfo struct {
	TargetID string `json:"targetId"`
	Type     string `json:"type"`
	Title    string `json:"title"`
	URL      string `json:"url"`
	Attached bool   `json:"attached"`
	OpenerId string `json:"openerId,omitempty"`
}
