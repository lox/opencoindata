package web

import (
	"net/http"
	"sync"
	"github.com/gorilla/websocket"
)

type WebSockets struct {
	clients map[string]*websocket.Conn
	mutex   sync.RWMutex
}

func (w *WebSockets) Add(conn *websocket.Conn) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if w.clients == nil {
		w.clients = map[string]*websocket.Conn{}
	}
	w.clients[conn.RemoteAddr().String()] = conn
}

func (w *WebSockets) Delete(conn *websocket.Conn) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	delete(w.clients, conn.RemoteAddr().String())
}

func (w *WebSockets) Map(callback func(conn *websocket.Conn)) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	for _, c := range w.clients {
		callback(c)
	}
}

// a http.HandlerFunc to serve a websocket
func WebSocketHandler(handler func(ws *websocket.Conn)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
		if _, ok := err.(websocket.HandshakeError); ok {
			http.Error(w, "Not a websocket handshake", 400)
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			handler(ws)
		}
	}
}
