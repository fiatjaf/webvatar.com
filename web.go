package main

import (
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"github.com/PuerkitoBio/goquery"
	"github.com/go-martini/martini"
	"github.com/hoisie/redis"
	"io"
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
		if strings.HasPrefix(target, "http://") == false &&
			strings.HasPrefix(target, "https://") == false {
			target = "http://" + target
		}
		t, err := url.Parse(target)
		if err != nil {
			log.Print(err)
			http.Error(res, "Oops!", http.StatusNotFound)
			return
		}

		// important variables ahead:
		target = t.String()
		domain := t.Host + t.Path
		// remove ending slash from domain
		if strings.HasSuffix(domain, "/") {
			domain = domain[:len(domain)-1]
		}

		// acceptsmall= or s= (default: y)
		acceptSmall := !(qs.Get("acceptsmall") == "n" || qs.Get("s") == "n")

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
			log.Print("redis got", cached)

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

			// if not acceptSmal, check if the image is small and send an alternative
			cachedImageSize := getImageSize(cached, insecureClient)
			if !acceptSmall && cachedImageSize < 4500 {
				log.Print("cached image is", cachedImageSize, "which is too small. send alternative.")
				imageDefault := alternative(d, domain)
				http.Redirect(res, req, imageDefault, 302)
			}

			// otherwise we send the cached url
			http.Redirect(res, req, cached, 302)
			return
		}

		// no cache or forcedefault, proceed to scan the images
		// we will fetch all images first
		best := bestImage{url: "", size: 0, tried: make(map[string]int)}

		// parse HTML in search for images
		htmlResp, err := insecureClient.Get(target)
		if err != nil {
			log.Print(err)
			// url is errored, save "404" to cache to avoid scanning it again every time
			rd.Setex(domain, 1296000, []byte("404"))
			http.Error(res, "Oops!", http.StatusNotFound)
			return
		}
		// update our base url with the url after redirection
		t = htmlResp.Request.URL

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
			if found == false || imageSource == "" {
				// inexistent
				return true
			}
			if _, ok := best.tried[imageSource]; ok {
				// duplicated
				return true
			}
			log.Print("inspecting " + imageSource)
			best.tried[imageSource] = 1
			uImageSource, err := url.Parse(imageSource)
			if err != nil {
				return true
			}
			imageUrl := t.ResolveReference(uImageSource).String()

			return handleImageUrl(
				imageUrl,
				&best,
				t,
				insecureClient,
			)
		})
		// look for rel=icon (href) -- but only if nothing sufficiently big was found yet
		if best.size < 8500 {
			log.Print("will search link[rel]")
			doc.Find("link[rel]").EachWithBreak(func(i int, s *goquery.Selection) bool {
				// reduce the list to only those who have "icon" or the apple thing
				rel, _ := s.Attr("rel")
				rels := strings.Fields(rel)
				potentialRels := []string{"icon", "apple-touch-icon-precomposed"}
				if !(findAny(rels, potentialRels)) {
					return true
				}

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
					&best,
					t,
					insecureClient,
				)
			})
		}

		// after testing the matches, verify the results
		if best.url == "" {
			// none found fallback to default
			best.url = alternative(d, domain)

			// save 204 (No Content) to cache
			rd.Setex(domain, 1296000, []byte("204"))
		} else {
			// if found, save url in cache
			rd.Setex(domain, 1296000, []byte(best.url))

			// if not acceptSmall and the image is small, send an alternative
			if !acceptSmall && best.size < 4500 {
				best.url = alternative(d, domain)
			}
		}
		log.Print("after image search, redirecting to", best)
		http.Redirect(res, req, best.url, 302)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}
	m.RunOnAddr("0.0.0.0:" + port)
}

type bestImage struct {
	url   string
	size  int
	tried map[string]int
}

func handleImageUrl(
	imageUrl string,
	best *bestImage,
	t *url.URL,
	insecureClient *http.Client,
) bool {

	// found image, test size
	size := getImageSize(imageUrl, insecureClient)
	if size == -1 {
		// getImageSize returns -1 when an error occurs,
		// in this case skip everything by returning true
		log.Print("error fetching image size.")
		return true
	}
	if size > best.size {
		best.url = imageUrl
		best.size = size

		// stop searching if image is reasonably big
		if size > 8500 {
			log.Print("sufficient size found,", size, "stopping")
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
		case "blank":
			return "https://secure.gravatar.com/avatar/webvatar.com?d=blank"
		case "robohash":
			return "http://robohash.org/" + domain
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
			return "http://chart.apis.google.com/chart?chst=d_text_outline&chld=666|42|h|000|_|||" + strings.Join(lines, "|") + "|"
		default:
			return "https://secure.gravatar.com/avatar/" + computeMD5(domain) + "?d=" + defaultImage
		}
	}
}

func getImageSize(url string, insecureClient *http.Client) int {
	imageResp, err := insecureClient.Head(url)
	if err != nil {
		return -1
	}
	sSize := imageResp.Header.Get("Content-Length")
	if sSize == "" {
		sSize = imageResp.Header.Get("content-length")
	}
	if sSize == "" {
		// if there's no content-length, assume it is reasonable
		return 4501
	}

	nSize, err := strconv.Atoi(sSize)
	if err != nil {
		log.Print(err)
		return -1
	}
	return nSize
}

func computeMD5(str string) string {
	h := md5.New()
	io.WriteString(h, str)
	return hex.EncodeToString(h.Sum(nil))
}

func findAny(list []string, what []string) bool {
	for _, n := range list {
		for _, w := range what {
			if w == n {
				return true
			}
		}
	}
	return false
}
