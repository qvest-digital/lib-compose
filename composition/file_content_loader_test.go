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

	loader := NewFileContentLoader(true, true)
	fd := NewFetchDefinition(FileURLPrefix + fileName)
	fd.Name = "content"
	c, err := loader.Load(fd)
	a.Equal("content", c.Name())
	assertContentLoaded(t, c, err, "some head content")
}

func Test_FileContentLoader_LoadIndexForDirectory(t *testing.T) {
	a := assert.New(t)

	dir := os.TempDir()
	fileName := filepath.Join(dir, "index.html")

	err := ioutil.WriteFile(fileName, []byte("<html><head>some head content</head></html>"), 0660)
	a.NoError(err)

	loader := NewFileContentLoader(true, true)
	c, err := loader.Load(NewFetchDefinition(FileURLPrefix + dir))
	assertContentLoaded(t, c, err, "some head content")
}

func assertContentLoaded(t *testing.T, c Content, err error, s string) {
	a := assert.New(t)
	a.NoError(err)
	a.NotNil(c)
	a.Nil(c.Reader())
	eqFragment(t, "some head content", c.Head())
	a.Equal(0, len(c.Body()))
}

func Test_FileContentLoader_LoadStream(t *testing.T) {
	a := assert.New(t)

	fileName := filepath.Join(os.TempDir(), randString(10)+".css")

	err := ioutil.WriteFile(fileName, []byte("some non html content"), 0660)
	a.NoError(err)

	loader := NewFileContentLoader(true, true)
	c, err := loader.Load(NewFetchDefinition(FileURLPrefix + fileName))
	a.NoError(err)
	a.NotNil(c)
	body, err := ioutil.ReadAll(c.Reader())
	a.NoError(err)
	a.Equal("some non html content", string(body))
}

func Test_FileContentLoader_LoadError(t *testing.T) {
	a := assert.New(t)

	// create a file without read permissions
	f, err := ioutil.TempFile("", "lib-compose-test")
	a.NoError(err)
	a.NoError(f.Chmod(os.FileMode(0)))
	f.Close()

	loader := NewFileContentLoader(true, true)
	_, err = loader.Load(NewFetchDefinition(f.Name()))
	a.Error(err)
}

func Test_FileContentLoader_Return404IfFileNotFound(t *testing.T) {
	a := assert.New(t)

	loader := NewFileContentLoader(true, true)
	c, err := loader.Load(NewFetchDefinition("/tmp/some/non/existing/path"))
	a.NotNil(c)
	a.Error(err)
	a.Equal(404, c.HttpStatusCode())
}

func Test_FileContentLoader_RequestProcessor(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fd := NewFetchDefinition("/tmp/some/non/existing/path")
	fd.RespProc = NewMockResponseProcessor(ctrl)

	_, err := NewFileContentLoader(true, true).Load(fd)
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
