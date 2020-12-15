package main

import (
	"net"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"golang.org/x/net/websocket"
)

var watchdog struct {
	*time.Ticker
	numActive  int32
	lastActive int64
}

func init() {
	watchdog.Ticker = time.NewTicker(time.Minute)
	watchdog.lastActive = time.Now().Add(time.Hour).UnixNano()

	go func() {
		for range watchdog.C {
			if atomic.LoadInt32(&watchdog.numActive) > 0 {
				continue
			}
			t := time.Unix(0, atomic.LoadInt64(&watchdog.lastActive))
			if time.Now().After(t.Add(time.Minute)) {
				shutdown <- os.Interrupt
			}
		}
	}()
}

func middlewareWatcher(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&watchdog.numActive, +1)
		defer atomic.AddInt32(&watchdog.numActive, -1)
		atomic.StoreInt64(&watchdog.lastActive, time.Now().UnixNano())
		next.ServeHTTP(w, r)
	})
}

func connectionWatcher(conn net.Conn, state http.ConnState) {
	switch state {
	case http.StateNew:
		atomic.AddInt32(&watchdog.numActive, +1)
	case http.StateHijacked, http.StateClosed:
		atomic.AddInt32(&watchdog.numActive, -1)
	case http.StateActive:
		atomic.StoreInt64(&watchdog.lastActive, time.Now().UnixNano())
	}
}

func websocketWatcher(conn *websocket.Conn) {
	websockets.Register(conn)
	defer websockets.Unregister(conn)

	for {
		if err := websockets.ReadPong(conn); err == nil {
			continue
		} else if !os.IsTimeout(err) {
			return
		}
		if err := websockets.SendPing(conn); err != nil {
			return
		}
	}
}
