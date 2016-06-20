package composition

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func Test_FileContentLoader_LoadHTML(t *testing.T) {
	a := assert.New(t)

	fileName := filepath.Join(os.TempDir(), randString(10)+".html")

	err := ioutil.WriteFile(fileName, []byte("<html><head>some head content</head></html>"), 0660)
	a.NoError(err)

	loader := NewFileContentLoader()
	c, err, _ := loader.Load(NewFetchDefinition(FileURLPrefix + fileName))
	a.NoError(err)
	a.NotNil(c)
	a.Nil(c.Reader())
	a.Equal(FileURLPrefix+fileName, c.URL())
	eqFragment(t, "some head content", c.Head())
	a.Equal(0, len(c.Body()))
}

func Test_FileContentLoader_LoadStream(t *testing.T) {
	a := assert.New(t)

	fileName := filepath.Join(os.TempDir(), randString(10)+".css")

	err := ioutil.WriteFile(fileName, []byte("some non html content"), 0660)
	a.NoError(err)

	loader := NewFileContentLoader()
	c, err, _ := loader.Load(NewFetchDefinition(FileURLPrefix + fileName))
	a.NoError(err)
	a.NotNil(c)
	body, err := ioutil.ReadAll(c.Reader())
	a.NoError(err)
	a.Equal("some non html content", string(body))
}

func Test_FileContentLoader_LoadError(t *testing.T) {
	a := assert.New(t)

	loader := NewFileContentLoader()
	_, err, _ := loader.Load(NewFetchDefinition("/tmp/some/non/existing/path"))
	a.Error(err)
}

func Test_FileContentLoader_RequestProcessor(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fd := NewFetchDefinition("/tmp/some/non/existing/path")
	fd.RespProc = NewMockResponseProcessor(ctrl)

	_, err, _ := NewFileContentLoader().Load(fd)
	a.Equal(ResponseProcessorsNotApplicable, err)
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
