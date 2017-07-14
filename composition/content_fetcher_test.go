package composition

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
	"time"
)

func Test_ContentFetcher_FetchingWithDependency(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	a := assert.New(t)

	loader := NewMockContentLoader(ctrl)
	barFd := getFetchDefinitionMock(ctrl, loader, "/bar", nil, time.Millisecond*2, map[string]interface{}{"foo": "bar"})
	fooFd := getFetchDefinitionMock(ctrl, loader, "/foo", []*FetchDefinition{barFd}, time.Millisecond*2, map[string]interface{}{"bli": "bla"})
	bazzFd := getFetchDefinitionMock(ctrl, loader, "/bazz", []*FetchDefinition{barFd}, time.Millisecond, map[string]interface{}{})

	fetcher := NewContentFetcher(nil, true, true)
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

	fetcher := NewContentFetcher(nil, true, true)
	fetcher.Loader = loader

	a.True(fetcher.Empty())
}

func Test_ContentFetcher_LazyDependencies(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	a := assert.New(t)

	loader := NewMockContentLoader(ctrl)

	parent := NewFetchDefinition("/parent")
	content := NewMockContent(ctrl)
	loader.EXPECT().
		Load(parent).
		Return(content, nil)

	content.EXPECT().
		RequiredContent().
		Return([]*FetchDefinition{})
	content.EXPECT().
		Meta().
		Return(nil)
	content.EXPECT().
		Dependencies().
		Return(map[string]Params{"child": Params{"foo": "bar"}})

	child := getFetchDefinitionMock(ctrl, loader, "/child", nil, time.Millisecond*2, nil)

	fetcher := NewContentFetcher(nil, true, true)
	fetcher.Loader = loader
	fetcher.SetFetchDefinitionFactory(func(name string, params Params) (fd *FetchDefinition, exist bool, err error) {
		a.Equal("child", name)
		a.Equal("bar", params["foo"])
		return child, true, nil
	})

	fetcher.AddFetchJob(parent)
	results := fetcher.WaitForResults()

	a.Equal(2, len(results))
	a.False(fetcher.Empty())
	a.Equal("/parent", results[0].Def.URL)
	a.Equal("/child", results[1].Def.URL)
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

	content.EXPECT().
		Dependencies().
		Return(map[string]Params{})

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

	barFd := NewFetchDefinition("/bar").WithPriority(30)
	fooFd := NewFetchDefinition("/foo").WithPriority(10)
	bazzFd := NewFetchDefinition("/bazz").WithPriority(5)

	results := []*FetchResult{{Def: barFd}, {Def: fooFd}, {Def: bazzFd}}

	a.Equal(30, results[0].Def.Priority)
	a.Equal(10, results[1].Def.Priority)
	a.Equal(5, results[2].Def.Priority)

	sort.Sort(FetchResults(results))

	a.Equal(5, results[0].Def.Priority)
	a.Equal(10, results[1].Def.Priority)
	a.Equal(30, results[2].Def.Priority)
}

func Test_ContentFetcher_PriorityOrderAfterFetchCompletion(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	a := assert.New(t)

	loader := NewMockContentLoader(ctrl)
	barFd := getFetchDefinitionMock(ctrl, loader, "/bar", nil, time.Millisecond*2, map[string]interface{}{"foo": "bar"})
	barFd.Priority = 1024
	fooFd := getFetchDefinitionMock(ctrl, loader, "/foo", nil, time.Millisecond*2, map[string]interface{}{"bli": "bla"})
	fooFd.Priority = 211
	bazzFd := getFetchDefinitionMock(ctrl, loader, "/bazz", nil, time.Millisecond, map[string]interface{}{})
	bazzFd.Priority = 412

	fetcher := NewContentFetcher(nil, true, true)
	fetcher.Loader = loader

	fetcher.AddFetchJob(barFd)
	fetcher.AddFetchJob(fooFd)
	fetcher.AddFetchJob(bazzFd)

	results := fetcher.WaitForResults()

	a.Equal(211, results[0].Def.Priority)
	a.Equal(412, results[1].Def.Priority)
	a.Equal(1024, results[2].Def.Priority)

}
