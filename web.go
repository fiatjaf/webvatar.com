package main

import (
	"crypto/tls"
	"github.com/PuerkitoBio/goquery"
	"github.com/go-martini/martini"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

func main() {
	m := martini.Classic()

	// custom http insecure client
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	insecureClient := &http.Client{Transport: tr}

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

		// we will fetch all images first
		bestImageUrl := ""
		bestImageSize := 0

		// parse HTML in search for images
		htmlResp, err := insecureClient.Get(t.String())
		if err != nil {
			log.Fatal(err)
			http.Error(res, "Oops!", http.StatusInternalServerError)
			return
		}
		doc, err := goquery.NewDocumentFromResponse(htmlResp)
		if err != nil {
			log.Fatal(err)
			http.Error(res, "Oops!", http.StatusInternalServerError)
			return
		}
		doc.Find("h-card u-photo, [rel=\"icon\"]").EachWithBreak(func(i int, s *goquery.Selection) bool {
			imageHref, found := s.Attr("href")
			if found == false {
				return true
			}
			uImageHref, err := url.Parse(imageHref)
			if err != nil {
				return true
			}
			imageUrl := t.ResolveReference(uImageHref).String()

			// found image, test size
			imageResp, err := insecureClient.Head(imageUrl)
			if err != nil {
				return true
			}
			sSize := imageResp.Header.Get("Content-Length")
			if sSize == "" {
				sSize = imageResp.Header.Get("content-length")
			}
			nSize, err := strconv.Atoi(sSize)
			if err != nil {
				return true
			}
			if nSize > bestImageSize {
				bestImageUrl = imageUrl
				bestImageSize = nSize

				// stop searching if image is reasonably big
				if nSize > 8500 {
					return false
				}
			}

			// go test other findings
			return true
		})

		if bestImageUrl == "" {
			bestImageUrl = "http://google.com/"
		}
		http.Redirect(res, req, bestImageUrl, 302)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}
	m.RunOnAddr("0.0.0.0:" + port)
}
