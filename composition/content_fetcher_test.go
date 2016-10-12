package composition

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
        "sort"
)

func Test_ContentFetcher_FetchingWithDependency(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	a := assert.New(t)

	loader := NewMockContentLoader(ctrl)
	barFd := getFetchDefinitionMock(ctrl, loader, "/bar", nil, time.Millisecond*2, map[string]interface{}{"foo": "bar"})
	fooFd := getFetchDefinitionMock(ctrl, loader, "/foo", []*FetchDefinition{barFd}, time.Millisecond*2, map[string]interface{}{"bli": "bla"})
	bazzFd := getFetchDefinitionMock(ctrl, loader, "/bazz", []*FetchDefinition{barFd}, time.Millisecond, map[string]interface{}{})

	fetcher := NewContentFetcher(nil)
	fetcher.Loader = loader

	fetcher.AddFetchJob(fooFd)
	fetcher.AddFetchJob(bazzFd)

	results := fetcher.WaitForResults()

	a.Equal(3, len(results))

	a.Equal("/foo", results[0].Def.URL)
	a.Equal("/bazz", results[1].Def.URL)
	a.Equal("/bar", results[2].Def.URL)

	meta := fetcher.MetaJSON()
	a.Equal("bar", meta["foo"])
	a.Equal("bla", meta["bli"])

	a.False(fetcher.Empty())
}

func Test_ContentFetcher_Empty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	a := assert.New(t)

	loader := NewMockContentLoader(ctrl)

	fetcher := NewContentFetcher(nil)
	fetcher.Loader = loader

	a.True(fetcher.Empty())
}

func getFetchDefinitionMock(ctrl *gomock.Controller, loaderMock *MockContentLoader, url string, requiredContent []*FetchDefinition, loaderBlocking time.Duration, metaJSON map[string]interface{}) *FetchDefinition {
	fd := NewFetchDefinition(url)
	fd.Timeout = time.Second * 42

	content := NewMockContent(ctrl)
	content.EXPECT().
		RequiredContent().
		Return(requiredContent)

	content.EXPECT().
		Meta().
		Return(metaJSON)

	loaderMock.EXPECT().
		Load(fd).
		Do(
			func(fetchDefinition *FetchDefinition) {
				time.Sleep(loaderBlocking)
			}).
		Return(content, nil)

	return fd
}

func Test_ContentFetchResultPrioritySort(t *testing.T) {
        a := assert.New(t)

        barFd := NewFetchDefinitionWithPriority("/bar", 30)
        fooFd := NewFetchDefinitionWithPriority("/foo", 10)
        bazzFd := NewFetchDefinitionWithPriority("/bazz", 5)

        results := []*FetchResult{{Def: barFd}, {Def: fooFd}, {Def: bazzFd}}

        a.Equal(30, results[0].Def.Priority)
        a.Equal(10, results[1].Def.Priority)
        a.Equal(5, results[2].Def.Priority)

        sort.Sort(FetchResults(results))

        a.Equal(5, results[0].Def.Priority)
        a.Equal(10, results[1].Def.Priority)
        a.Equal(30, results[2].Def.Priority)
}