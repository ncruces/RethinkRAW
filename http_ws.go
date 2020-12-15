package main

import (
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

var websockets = websocketManager{
	conns: make(map[*websocket.Conn]struct{}),
}

type websocketManager struct {
	sync.RWMutex
	conns map[*websocket.Conn]struct{}
}

func (ws *websocketManager) Register(conn *websocket.Conn) {
	ws.Lock()
	defer ws.Unlock()
	ws.conns[conn] = struct{}{}
}

func (ws *websocketManager) Unregister(conn *websocket.Conn) {
	ws.Lock()
	defer ws.Unlock()
	delete(ws.conns, conn)
}

func (ws *websocketManager) SendPing(conn *websocket.Conn) error {
	ws.RLock()
	defer ws.RUnlock()
	conn.PayloadType = websocket.PingFrame
	conn.SetWriteDeadline(time.Now().Add(time.Second))
	_, err := conn.Write(nil)
	return err
}

func (ws *websocketManager) ReadPong(conn *websocket.Conn) error {
	var dummy [8]byte
	conn.SetReadDeadline(time.Now().Add(time.Minute))
	_, err := conn.Read(dummy[:])
	return err
}

func (ws *websocketManager) Broadcast(message string) {
	ws.Lock()
	defer ws.Unlock()
	for conn := range ws.conns {
		conn.PayloadType = websocket.TextFrame
		conn.Write([]byte(message))
	}
}
