// Package chrome provides support to locate and run Google Chrome (or Microsoft Edge).
package chrome

import (
	"bufio"
	"bytes"
	"errors"
	"io/fs"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/ncruces/jason"
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
	url string
	msg atomic.Uint32
}

// Command returns the Cmd struct to execute a Chrome app loaded from url,
// and with the given user data and disk cache directories.
func Command(url string, dataDir, cacheDir string, arg ...string) *Cmd {
	once.Do(findChrome)
	if chrome == "" {
		return nil
	}

	prefs := filepath.Join(dataDir, "Default", "Preferences")
	if _, err := os.Stat(prefs); errors.Is(err, fs.ErrNotExist) {
		if err := os.MkdirAll(filepath.Dir(prefs), 0700); err == nil {
			os.WriteFile(prefs, []byte(`{
				"profile": {"cookie_controls_mode": 1},
				"search":  {"suggest_enabled": false},
				"signin":  {"allowed_on_next_startup": false},
				"enable_do_not_track": true
			}`), 0600)
		}
	}

	// https://peter.sh/experiments/chromium-command-line-switches/
	// https://github.com/GoogleChrome/chrome-launcher/blob/master/docs/chrome-flags-for-tools.md
	// https://source.chromium.org/chromium/chromium/src/+/master:chrome/test/chromedriver/chrome_launcher.cc
	arg = append([]string{
		"--app=" + url,
		"--user-data-dir=" + dataDir,
		"--disk-cache-dir=" + cacheDir,
		"--class=" + strconv.FormatUint(rand.Uint64(), 16),
		"--incognito", "--inprivate", "--bwsi", "--remote-debugging-port=0",
		"--no-first-run", "--no-default-browser-check", "--no-service-autorun", "--no-pings",
		"--disable-sync", "--disable-breakpad", "--disable-extensions", "--disable-default-apps",
		"--disable-component-extensions-with-background-pages", "--disable-background-networking",
		"--disable-domain-reliability", "--disable-client-side-phishing-detection", "--disable-component-update",
	}, arg...)

	cmd := exec.Command(chrome, arg...)
	return &Cmd{cmd: cmd, url: origin(url)}
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
	pipe, err := c.cmd.StderrPipe()
	if err != nil {
		return err
	}
	defer pipe.Close()

	err = c.cmd.Start()
	if err != nil {
		return err
	}

	scan := bufio.NewScanner(pipe)
	for scan.Scan() {
		const prefix = "DevTools listening on "
		line := scan.Bytes()
		if bytes.HasPrefix(line, []byte(prefix)) {
			url := line[len(prefix):]
			c.ws, _, err = websocket.DefaultDialer.Dial(string(url), nil)
			if err != nil {
				return err
			}
			go c.receiveloop()
			return nil
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

func (c *Cmd) receiveloop() {
	var started bool
	targets := set[string]{}
	c.send("Target.setDiscoverTargets", "", jason.Object{"discover": true})
	for {
		var msg cdpMessage
		err := c.ws.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseNormalClosure,
				websocket.CloseAbnormalClosure) {
				log.Println("chrome:", err)
			}
			break
		}
		switch msg.Method {
		case "Target.targetDestroyed", "Target.targetCrashed":
			targets.Del(jason.ToA[string](msg.Params["targetId"]))
		case "Target.targetCreated", "Target.targetInfoChanged":
			info := jason.ToA[cdpTargetInfo](msg.Params["targetInfo"])
			if origin(info.URL) == c.url {
				targets.Add(info.TargetID)
			} else {
				targets.Del(info.TargetID)
			}
			if info.Type == "page" {
				started = true
			}
		}
		if started && len(targets) == 0 {
			c.Close()
		}
	}
}

func (c *Cmd) send(method, session string, params any) error {
	return c.ws.WriteJSON(jason.Object{
		"id":        c.msg.Add(1),
		"method":    method,
		"params":    params,
		"sessionId": session,
	})
}

type cdpMessage struct {
	ID     uint32          `json:"id,omitempty"`
	Method string          `json:"method,omitempty"`
	Result jason.RawValue  `json:"result,omitempty"`
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
