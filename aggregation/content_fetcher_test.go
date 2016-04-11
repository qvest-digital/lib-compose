package aggregation

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_ContentFetcher_Fetching(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	a := assert.New(t)

	fetcher := NewContentFetcher()

	d := NewFetchDefinition("/foo")

	content := NewMockContent(ctrl)
	content.EXPECT().RequiredContent().
		Return(nil)

	fetcher.contentLoaderForDefinition = func(d *FetchDefinition) ContentLoader {
		loader := NewMockContentLoader(ctrl)
		loader.EXPECT().Load(d.URL, d.Timeout).
			Return(content, nil)
		return loader
	}

	fetcher.AddFetchJob(d)
	results := fetcher.WaitForResults()

	a.Equal(1, len(results))
}
