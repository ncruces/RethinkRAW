package httpwatcher

import (
	"net"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"golang.org/x/net/websocket"
)

type Watcher struct {
	wsManager websocketManager
	wsServer  websocket.Server
	wsPath    string

	closed   chan struct{}
	active   chan struct{}
	numAlive int32

	handler   http.Handler
	connState func(conn net.Conn, state http.ConnState)
}

func NewWatcher(server *http.Server, websocketPath string, timeout time.Duration, onTimeout func()) *Watcher {
	var wd Watcher
	wd.closed = make(chan struct{})
	wd.active = make(chan struct{})
	wd.wsManager.conns = make(map[*websocket.Conn]struct{})

	wd.handler = server.Handler
	wd.connState = server.ConnState
	wd.wsServer.Handler = wd.websocketWatcher
	wd.wsPath = websocketPath

	server.ConnState = wd.connectionWatcher
	server.Handler = http.HandlerFunc(wd.middlewareWatcher)
	if wd.handler == nil {
		wd.handler = http.DefaultServeMux
	}

	go func() {
		for {
			timer := time.NewTimer(timeout)
			select {
			case <-wd.active:
				timer.Stop()
				continue
			case <-wd.closed:
				timer.Stop()
				return
			case <-timer.C:
				break
			}

			if atomic.LoadInt32(&wd.numAlive) > 0 {
				continue
			}
			if onTimeout != nil {
				onTimeout()
			}
			return
		}
	}()

	return &wd
}

func (wd *Watcher) middlewareWatcher(w http.ResponseWriter, r *http.Request) {
	wd.active <- struct{}{}
	atomic.AddInt32(&wd.numAlive, +1)
	defer atomic.AddInt32(&wd.numAlive, -1)

	if wd.wsPath != "" && wd.wsPath == r.URL.Path {
		wd.wsServer.ServeHTTP(w, r)
	} else {
		wd.handler.ServeHTTP(w, r)
	}
}

func (wd *Watcher) connectionWatcher(conn net.Conn, state http.ConnState) {
	switch state {
	case http.StateNew:
		atomic.AddInt32(&wd.numAlive, +1)
	case http.StateHijacked, http.StateClosed:
		atomic.AddInt32(&wd.numAlive, -1)
	case http.StateActive:
		wd.active <- struct{}{}
	}
	if wd.connState != nil {
		wd.connState(conn, state)
	}
}

func (wd *Watcher) websocketWatcher(conn *websocket.Conn) {
	wd.wsManager.register(conn)
	defer wd.wsManager.unregister(conn)

	for {
		if err := wd.wsManager.readPong(conn); err == nil {
			continue
		} else if !os.IsTimeout(err) {
			return
		}
		if err := wd.wsManager.sendPing(conn); err != nil {
			return
		}
	}
}

func (wd *Watcher) Broadcast(msg string) {
	wd.wsManager.broadcast(msg)
}

func (wd *Watcher) Close() {
	wd.closed <- struct{}{}
}
