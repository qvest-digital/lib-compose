package main

import (
	"github.com/tarent/lib-compose/composition"
	"net/http"
)

type LazyFdFactory struct {
	r *http.Request
}

func NewLazyFdFactory(r *http.Request) *LazyFdFactory {
	return &LazyFdFactory{r}
}

func (fact *LazyFdFactory) getFetchDefinitions(name string, params composition.Params) (fd *composition.FetchDefinition, exist bool, err error) {
	baseUrl := "http://" + fact.r.Host
	if name == "teaser" {
		fd := composition.NewFetchDefinition(baseUrl + "/teaser?teaserId=foo" + params["teaserId"]).WithName("teaser")
		return fd, true, nil
	}
	return nil, false, nil
}
