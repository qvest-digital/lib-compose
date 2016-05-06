package main

import (
	gorilla "github.com/gorilla/handlers"
	"lib-ui-service/composition"
	"net/http"
	"os"
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
		fetcher := composition.NewContentFetcher()
		fetcher.AddFetchJob(composition.NewFetchDefinition("http://127.0.0.1:8080/static/styles.html"))
		fetcher.AddFetchJob(composition.NewFetchDefinition("http://127.0.0.1:8080/static/layout.html"))
		fetcher.AddFetchJob(composition.NewFetchDefinition("http://127.0.0.1:8080/static/lorem.html"))
		return fetcher
	}
	return composition.NewCompositionHandler(contentFetcherFactory)
}

func staticHandler() http.Handler {
	return http.FileServer(http.Dir("./"))
}
