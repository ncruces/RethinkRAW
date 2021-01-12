package httpwatcher

import (
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

type websocketManager struct {
	sync.RWMutex
	conns map[*websocket.Conn]struct{}
}

func (ws *websocketManager) register(conn *websocket.Conn) {
	ws.Lock()
	defer ws.Unlock()
	ws.conns[conn] = struct{}{}
}

func (ws *websocketManager) unregister(conn *websocket.Conn) {
	ws.Lock()
	defer ws.Unlock()
	delete(ws.conns, conn)
}

func (ws *websocketManager) sendPing(conn *websocket.Conn) error {
	ws.RLock()
	defer ws.RUnlock()
	conn.PayloadType = websocket.PingFrame
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	_, err := conn.Write(nil)
	return err
}

func (ws *websocketManager) readPong(conn *websocket.Conn) error {
	var dummy [8]byte
	conn.SetReadDeadline(time.Now().Add(time.Minute))
	_, err := conn.Read(dummy[:])
	return err
}

func (ws *websocketManager) broadcast(msg string) {
	ws.Lock()
	defer ws.Unlock()
	for conn := range ws.conns {
		conn.PayloadType = websocket.TextFrame
		conn.Write([]byte(msg))
	}
}
