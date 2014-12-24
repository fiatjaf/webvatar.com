package main

import (
	"github.com/go-martini/martini"
	"net/http"
	"os"
	"strings"
)

func main() {
	m := martini.Classic()

	m.Get("/", func(res http.ResponseWriter, req *http.Request) {
		http.Redirect(res, req, "http://indiewebcamp.com/webvatar", 302)
	})

	m.Get("/**", func(params martini.Params, res http.ResponseWriter, req *http.Request) {
		url := strings.ToLower(params["_1"])
		http.Redirect(res, req, url, 302)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}
	m.RunOnAddr("0.0.0.0:" + port)
}
