package main

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/go-martini/martini"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func main() {
	m := martini.Classic()

	m.Get("/", func(res http.ResponseWriter, req *http.Request) {
		http.Redirect(res, req, "http://indiewebcamp.com/webvatar", 302)
	})

	m.Get("/**", func(params martini.Params, res http.ResponseWriter, req *http.Request) {

		// parse url
		target := strings.ToLower(params["_1"])
		if strings.HasPrefix(target, "http://") == false || strings.HasPrefix(target, "https://") {
			target = "http://" + target
		}
		t, err := url.Parse(target)
		if err != nil {
			log.Fatal(err)
			http.Error(res, "Oops!", http.StatusInternalServerError)
			return
		}

		// parse HTML
		doc, err := goquery.NewDocument(t.String())
		if err != nil {
			log.Fatal(err)
			http.Error(res, "Oops!", http.StatusInternalServerError)
			return
		}
		imageUrl := ""
		doc.Find("h-card u-photo, [rel=\"icon\"]").Each(func(i int, s *goquery.Selection) {
			imageHref, found := s.Attr("href")
			if found == false {
				return
			}
			uImageHref, err := url.Parse(imageHref)
			if err != nil {
				log.Fatal(err)
				http.Error(res, "Oops!", http.StatusInternalServerError)
				return
			}
			imageUrl = t.ResolveReference(uImageHref).String()
			return
		})

		http.Redirect(res, req, imageUrl, 302)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}
	m.RunOnAddr("0.0.0.0:" + port)
}
