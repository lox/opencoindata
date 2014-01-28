package web

import (
	"log"
	"net/http"
	"sync"
	"github.com/davecgh/go-spew/spew"
	. "github.com/lox/opencoindata/core"

	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
	"github.com/gorilla/websocket"
)

type WebServer struct {
	clients map[string]*websocket.Conn
	mutex   sync.RWMutex
}

func NewWebServer() *WebServer {
	return &WebServer{clients: map[string]*websocket.Conn{}}
}

func (w *WebServer) addClient(conn *websocket.Conn) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.clients[conn.RemoteAddr().String()] = conn
}

func (w *WebServer) deleteClient(conn *websocket.Conn) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	delete(w.clients, conn.RemoteAddr().String())
}

func (w *WebServer) Serve(bind string) error {
	go NewTradeServer(func(t Trade) {
		w.mutex.Lock()
		defer w.mutex.Unlock()

		// log.Printf("Writing trade to %d clients", len(w.clients))
		for _, client := range w.clients {
			if err := client.WriteJSON(t); err != nil {
				panic(err)
			}
		}
	})

	m := martini.Classic()
	m.Use(martini.Static("assets"))
	m.Use(render.Renderer(render.Options{Directory: "./templates"}))
	m.Use(render.Renderer(render.Options{
		IndentJSON: true,
	}))

	m.Get("/", func(r render.Render) {
		r.HTML(200, "index", nil)
	})

	m.Get("/trades", func(r render.Render) {
		r.HTML(200, "trades", nil)
	})

	m.Get("/api/status", func(r render.Render) {

		status, err := GetPairStatus()
		spew.Dump(status)
		spew.Dump(err)

		/*
			var tables []struct {
				Name string
			}

			_, err := dbmap.Select(&tables, "SHOW TABLES")
			log.Printf("Error listing tables: %s", err.Error())

			log.Println("All rows:")
			for _, p := range tables {
				log.Printf("%v", p)
			}
		*/

		r.JSON(200, map[string]interface{}{"hello": "world"})
	})

	m.Get("/api/ws/trades", WebSocketHandler(func(ws *websocket.Conn) {
		w.addClient(ws)

		for {
			var obj struct{}
			err := ws.ReadJSON(&obj)
			if err != nil {
				w.deleteClient(ws)
				ws.Close()
				break
			} else {
				log.Printf("Client sent %v", obj)
			}
		}
	}))

	log.Printf("Listening on http://%s", bind)
	return http.ListenAndServe(bind, m)
}

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
