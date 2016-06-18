package main

import (
	gorilla "github.com/gorilla/handlers"
	"net/http"
	"os"
	"github.com/tarent/lib-composition/composition"
)

func main() {
	panic(http.ListenAndServe(":8080", handler()))
}

func handler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/static/", staticHandler())
	mux.Handle("/", compositionHandler())
	return gorilla.LoggingHandler(os.Stdout, mux)
}

func compositionHandler() http.Handler {
	contentFetcherFactory := func(r *http.Request) composition.FetchResultSupplier {
		defaultMetaJSON := map[string]interface{}{
			"header_text": "Hello World!",
			"request":     composition.MetadataForRequest(r),
		}

		fetcher := composition.NewContentFetcher(defaultMetaJSON)
		fetcher.AddFetchJob(composition.NewFetchDefinition("§[request.base_url]§/static/styles.html"))
		fetcher.AddFetchJob(composition.NewFetchDefinition("§[request.base_url]§/static/layout.html"))
		fetcher.AddFetchJob(composition.NewFetchDefinition("§[request.base_url]§/static/lorem.html"))
		return fetcher
	}
	return composition.NewCompositionHandler(contentFetcherFactory)
}

func staticHandler() http.Handler {
	return http.FileServer(http.Dir("./"))
}
