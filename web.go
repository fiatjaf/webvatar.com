package main

import (
	"crypto/tls"
	"github.com/PuerkitoBio/goquery"
	"github.com/go-martini/martini"
	"github.com/hoisie/redis"
	"log"
	"math"
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

	// connect to redis
	redisU, err := url.Parse(os.Getenv("REDISCLOUD_URL"))
	if err != nil {
		log.Fatal(err)
	}
	redisPw, _ := redisU.User.Password()
	var rd redis.Client
	rd.Addr = redisU.Host
	rd.Password = redisPw

	m.Get("/", func(res http.ResponseWriter, req *http.Request) {
		http.Redirect(res, req, "http://indiewebcamp.com/webvatar", 302)
	})

	m.Get("/**", func(params martini.Params, res http.ResponseWriter, req *http.Request) {
		log.Print("matched correct route")
		log.Print(params["_1"])
		qs := req.URL.Query()

		// parse url
		target := strings.ToLower(params["_1"])
		if strings.HasPrefix(target, "http://") == false ||
			strings.HasPrefix(target, "https://") == false {
			target = "http://" + target
		}
		t, err := url.Parse(target)
		if err != nil {
			log.Print(err)
			http.Error(res, "Oops!", http.StatusNotFound)
			return
		}
		target = t.String()
		domain := t.Host + t.Path
		// remove ending slash from domain
		if strings.HasSuffix(domain, "/") {
			domain = domain[:len(domain)-1]
		}

		// d= or alt=, defaultImage
		d := strings.Trim(qs.Get("d"), " ")
		if d == "" {
			d = strings.Trim(qs.Get("alt"), " ")
		}

		// if forcedefault, don't scan anything, just send the default
		if qs.Get("f") == "y" || qs.Get("forcedefault") == "y" {
			log.Print("--forcedefault")
			imageDefault := alternative(d, domain)
			http.Redirect(res, req, imageDefault, 302)
			return
		}

		// search url in cache
		cachedB, err := rd.Get(domain)
		if err == nil {
			cached := string(cachedB)
			log.Print("redis got")
			log.Print(cached)

			// redis cache can return a "204" string, in which case we use the default
			if cached == "204" {
				imageDefault := alternative(d, domain)
				http.Redirect(res, req, imageDefault, 302)
				return
			}

			// redis cache can also return a "404" string, meaning we shouldn't try to
			// extract images from the url, as it doesn't exist
			if cached == "404" {
				http.Error(res, "Oops!", http.StatusNotFound)
				return
			}

			// otherwise we send the cached url
			http.Redirect(res, req, cached, 302)
			return
		}

		// no cache or forcedefault, proceed to scan the images
		// we will fetch all images first
		bestImageUrl := ""
		bestImageSize := 0

		// parse HTML in search for images
		htmlResp, err := insecureClient.Get(target)
		if err != nil {
			log.Print(err)
			// url is errored, save "404" to cache to avoid scanning it again every time
			rd.Setex(domain, 1296000, []byte("404"))
			http.Error(res, "Oops!", http.StatusNotFound)
			return
		}
		doc, err := goquery.NewDocumentFromResponse(htmlResp)
		if err != nil {
			log.Print(err)
			// document is not html, save "204" to cache and return alternative
			rd.Setex(domain, 1296000, []byte("204"))
			imageDefault := alternative(d, domain)
			http.Redirect(res, req, imageDefault, 302)
			return
		}
		// look for h-card pictures (src)
		doc.Find(".h-card .u-photo").EachWithBreak(func(i int, s *goquery.Selection) bool {
			imageSource, found := s.Attr("src")
			if found == false {
				return true
			}
			log.Print("inspecting " + imageSource)
			uImageSource, err := url.Parse(imageSource)
			if err != nil {
				return true
			}
			imageUrl := t.ResolveReference(uImageSource).String()

			return handleImageUrl(
				imageUrl,
				bestImageUrl,
				bestImageSize,
				qs.Get("acceptsmall") == "y" || qs.Get("s") == "y",
				t,
				insecureClient,
			)
		})
		// look for rel=icon (href)
		doc.Find(".h-card .u-photo").EachWithBreak(func(i int, s *goquery.Selection) bool {
			imageSource, found := s.Attr("href")
			if found == false {
				return true
			}
			log.Print("inspecting " + imageSource)
			uImageSource, err := url.Parse(imageSource)
			if err != nil {
				return true
			}
			imageUrl := t.ResolveReference(uImageSource).String()

			return handleImageUrl(
				imageUrl,
				bestImageUrl,
				bestImageSize,
				qs.Get("acceptsmall") == "y" || qs.Get("s") == "y",
				t,
				insecureClient,
			)
		})

		// after testing the matches, verify the results
		if bestImageUrl == "" {
			// none found fallback to default
			bestImageUrl = alternative(d, domain)

			// save 204 (No Content) to cache
			rd.Setex(domain, 1296000, []byte("204"))
		} else {
			// if found, save url in cache
			rd.Setex(domain, 1296000, []byte(bestImageUrl))
		}
		log.Print("after image search, redirecting to")
		log.Print(bestImageUrl)
		http.Redirect(res, req, bestImageUrl, 302)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}
	m.RunOnAddr("0.0.0.0:" + port)
}

func handleImageUrl(
	imageUrl string,
	bestImageUrl string,
	bestImageSize int,
	acceptSmall bool,
	t *url.URL,
	insecureClient *http.Client,
) bool {

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

	// minimum size threshold
	if !acceptSmall && nSize < 4500 {
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
}

func alternative(defaultImage string, domain string) string {
	if strings.HasPrefix(defaultImage, "http") {
		return defaultImage
	} else {
		if defaultImage == "" {
			defaultImage = "nameshow"
		}
		switch defaultImage {
		case "robohash":
			return "http://robohash.org/" + domain
		case "blank":
			return "https://secure.gravatar.com/avatar/webvatar.com?d=blank"
		case "nameshow":
			linelen := len(domain)
			var n float64
			if linelen > 30 {
				n = float64(linelen) / 3
			} else if float64(linelen) > 20 {
				n = float64(linelen) / 2
			} else {
				n = float64(linelen)
			}
			n = math.Ceil(n)
			nabs := int(n)
			lines := make([]string, 0)
			for j := 0; j <= linelen; j += nabs {
				var part string
				if j+nabs <= linelen {
					part = domain[j : j+nabs]
				} else {
					part = domain[j:linelen]
				}
				lines = append(lines, part)
			}
			return "http://chart.apis.google.com/chart?chst=d_text_outline&chld=666|42|h|000|_|||" + strings.Join(lines, "|") + "||"
		default:
			return "https://secure.gravatar.com/avatar/" + domain + "?d=" + defaultImage
		}
	}
}
