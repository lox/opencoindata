package web

import (
	"log"
	"net/http"

	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
)

func Run(bind string) error {
	m := martini.Classic()
	m.Use(martini.Static("assets"))
	m.Use(render.Renderer(render.Options{Directory: "./templates"}))

	m.Get("/", func(r render.Render) {
		r.HTML(200, "index", nil)
	})

	log.Printf("Listening on http://%s", bind)
	return http.ListenAndServe(bind, m)
}
