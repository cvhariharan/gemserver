package main

import (
	"io/fs"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/blevesearch/bleve/v2"
	gemini "github.com/cvhariharan/gemini-server"
	"github.com/patrickmn/go-cache"
)

var indexAlias bleve.IndexAlias

// RATE_LIMIT sets the number of requests per minute
const RATE_LIMIT = 15

func main() {
	mountPoint := os.Getenv("AWS_EFS_MOUNT")
	if mountPoint == "" {
		log.Fatal("mount point no set. AWS_EFS_MOUNT empty")
	}

	indexAlias = bleve.NewIndexAlias()
	filepath.Walk(mountPoint, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() && strings.HasPrefix(info.Name(), "index-") {
			log.Println(path)
			if err != nil {
				return err
			}

			index, err := bleve.Open(path)
			if err != nil {
				log.Println(err)
				return err
			}
			indexAlias.Add(index)
		}
		return nil
	})
	defer indexAlias.Close()

	c := cache.New(60*time.Second, 5*time.Minute)

	gemini.Handle("/search", rateLimit(gemini.Handlerfunc(searchHandler), c))

	gemini.Handle("/", rateLimit(gemini.StripPrefix("/static", gemini.FileServer("./static/index.gmi")), c))

	log.Fatal(gemini.ListenAndServeTLS(":1965", "gemini.crt", "gemini.key"))
}

func searchHandler(w *gemini.Response, r *gemini.Request) {
	if len(r.URL.RawQuery) == 0 {
		w.SetStatus(gemini.StatusInput, "Search")
		w.Write([]byte("# Search Engine"))
	} else {
		queryStr, err := url.QueryUnescape(r.URL.RawQuery)
		if err != nil {
			w.SetStatus(gemini.StatusTemporaryFailure, "Could not decode search query")
			w.SendStatus()
		}
		log.Println(queryStr)

		query := bleve.NewQueryStringQuery(queryStr)
		searchRequest := bleve.NewSearchRequest(query)
		searchResult, _ := indexAlias.Search(searchRequest)
		log.Println(searchResult)
		var links string
		if len(searchResult.Hits) > 0 {
			for _, v := range searchResult.Hits {
				links = links + "=>" + v.ID + " " + v.ID + "\n"
			}
		}
		w.Write([]byte(links))
	}
}

func rateLimit(h gemini.Handler, c *cache.Cache) gemini.Handler {
	return gemini.Handlerfunc(func(w *gemini.Response, r *gemini.Request) {
		ipPort := strings.Split(w.Body.RemoteAddr().String(), ":")
		ip := strings.Join(ipPort[:len(ipPort)-1], ":")

		val, ok := c.Get(ip)

		if !ok {
			c.Set(ip, RATE_LIMIT, cache.DefaultExpiration)
		} else {
			reqNo := val.(int)
			if reqNo <= 0 {
				w.SetStatus(gemini.StatusTemporaryFailure, "Too many requests")
				w.SendStatus()
				return
			}
			c.Decrement(ip, 1)
		}
		h.ServeGemini(w, r)
	})
}
