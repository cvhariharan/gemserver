package main

import (
	"log"
	"net/url"
	"os"
	"path/filepath"

	"github.com/blevesearch/bleve/v2"
	gemini "github.com/cvhariharan/gemini-server"
)

const INDEX = "index"

var index bleve.Index

func main() {
	mountPoint := os.Getenv("AWS_EFS_MOUNT")
	if mountPoint == "" {
		log.Fatal("mount point no set. AWS_EFS_MOUNT empty")
	}

	index, _ = bleve.Open(filepath.Join(mountPoint, INDEX))

	gemini.HandleFunc("/search", searchHandler)

	gemini.HandleFunc("/", landingPage)

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
		searchResult, _ := index.Search(searchRequest)
		var links string
		if len(searchResult.Hits) > 0 {
			for _, v := range searchResult.Hits {
				links = links + "=>" + v.ID + " " + v.ID + "\n"
			}
		}
		w.Write([]byte(links))
	}
}

func landingPage(w *gemini.Response, r *gemini.Request) {
	w.SetStatus(gemini.StatusSuccess, "text/gemini")
	resp := "# Gemini Search Engine\n=> /search Search\n=> https://blevesearch.com/docs/Query-String-Query/ Query Guide\n=> https://github.com/cvhariharan/gemini-crawler Crawler Repo\n=> https://github.com/cvhariharan/gemsearch Server Repo\n"
	w.Write([]byte(resp))
}
