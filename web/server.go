package web

import (
	"log"
	"net/http"
	. "github.com/lox/opencoindata/core"

	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
	"github.com/gorilla/websocket"
)

// configuration for serving the web app
type ServeConfig struct {
	// the main host and port to listen on for requests
	BindAddress string
	// the hostname to serve api requests from, by default any
	ApiHostname string
	// the hostname to serve websockets from, by default any
	WsHostname string
}

// listens for http requests and routes them to website, api and websockets
func Serve(config ServeConfig) error {
	if config.ApiHostname != "" {
		log.Printf("Serving api on hostname %s", config.ApiHostname)
	}
	if config.WsHostname != "" {
		log.Printf("Serving websockets on hostname %s", config.WsHostname)
	}

	http.Handle(config.ApiHostname+"/api/", serveApi(config))
	http.Handle(config.WsHostname+"/ws/", serveWebsockets(config))
	http.Handle("/", serveWebsite(config))

	log.Printf("Listening on http://" + config.BindAddress)
	return http.ListenAndServe(config.BindAddress, nil)
}

// handles requests to the user-facing website
func serveWebsite(config ServeConfig) http.Handler {
	m := martini.Classic()
	m.Use(martini.Static("assets"))
	m.Use(render.Renderer(render.Options{
		Directory: "./templates",
		Layout:    "base",
	}))

	m.Get("/", func(r render.Render) {
		r.HTML(200, "index", nil)
	})

	m.Get("/trades", func(r render.Render) {
		r.HTML(200, "trades", map[string]interface{}{
			"WsHostname": config.WsHostname,
		})
	})

	m.Get("/status", func(r render.Render) error {
		status, err := GetAllPairStatus()
		if err != nil {
			return err
		}

		r.HTML(200, "status", map[string]interface{}{"Status": status})
		return nil
	})

	return m
}

// handles requests to the restful api
func serveApi(config ServeConfig) http.Handler {
	m := martini.Classic()
	m.Use(render.Renderer(render.Options{
		IndentJSON: true,
	}))

	m.Get("/api/v1/status", func(r render.Render) {
		status, err := GetAllPairStatus()
		if err != nil {
			r.JSON(500, map[string]interface{}{"error": err.Error()})
			return
		}

		r.JSON(200, status)
	})

	return m
}

// serves the real-time websocket api
func serveWebsockets(config ServeConfig) http.Handler {
	var clients WebSockets

	// fire up a trade server to listen for trades from the collector
	go NewTradeServer(func(t Trade) {
		clients.Map(func(client *websocket.Conn) {
			if err := client.WriteJSON(t); err != nil {
				panic(err)
			}
		})
	})

	m := martini.Classic()

	// a route for websockets to connect to be sent trades
	m.Get("/ws/v1/trades", WebSocketHandler(func(ws *websocket.Conn) {
		clients.Add(ws)

		for {
			var obj struct{}
			err := ws.ReadJSON(&obj)
			if err != nil {
				clients.Delete(ws)
				ws.Close()
				break
			} else {
				log.Printf("Client sent %v", obj)
			}
		}
	}))

	return m
}
