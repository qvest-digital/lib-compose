package logging

import (
	"bytes"
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

type logReccord struct {
	Type              string            `json:"type"`
	Timestamp         string            `json:"@timestamp"`
	CorrelationId     string            `json:"correlation_id"`
	RemoteIp          string            `json:"remote_ip"`
	Host              string            `json:"host"`
	URL               string            `json:"url"`
	Method            string            `json:"method"`
	Duration          int               `json:"duration"`
	ResponseStatus    int               `json:"response_status"`
	UserCorrelationId string            `json:"user_correlation_id"`
	Cookies           map[string]string `json:"cookies"`
	Error             string            `json:"error"`
	Message           string            `json:"message"`
}

func Test_Logger_Set(t *testing.T) {
	a := assert.New(t)

	// given: an error logger in text format
	Set("error", true)
	defer Set("info", false)
	Logger.Formatter.(*logrus.TextFormatter).DisableColors = true
	b := bytes.NewBuffer(nil)
	Logger.Out = b

	// when: I log something
	Logger.Info("should be ignored ..")
	Logger.WithField("foo", "bar").Error("oops")

	// then: only the error text is contained
	// and it is text formated
	a.Regexp(`^time.* level\=error msg\=oops foo\=bar.*`, b.String())
}

func Test_Logger_Access(t *testing.T) {
	a := assert.New(t)

	// given a logger
	b := bytes.NewBuffer(nil)
	Logger.Out = b
	AccessLogCookiesBlacklist = []string{"ignore", "user_id"}
	UserCorrelationCookie = "user_id"

	// and a request
	r, _ := http.NewRequest("GET", "http://www.example.org/foo?q=bar", nil)
	r.Header = http.Header{
		CorrelationIdHeader: {"correlation-123"},
		"Cookie":            {"user_id=user-id-xyz; ignore=me; foo=bar;"},
	}
	r.RemoteAddr = "127.0.0.1"

	// when: We log a request with access
	start := time.Now().Add(-1 * time.Second)
	Access(r, start, 201)

	// then: all fields match
	data := &logReccord{}
	err := json.Unmarshal(b.Bytes(), data)
	a.NoError(err)

	a.Equal(map[string]string{"foo": "bar"}, data.Cookies)
	a.Equal("correlation-123", data.CorrelationId)
	a.InDelta(1000, data.Duration, 0.5)
	a.Equal("", data.Error)
	a.Equal("www.example.org", data.Host)
	a.Equal("GET", data.Method)
	a.Equal("201 GET /foo?...", data.Message)
	a.Equal("127.0.0.1", data.RemoteIp)
	a.Equal(201, data.ResponseStatus)
	a.Equal("access", data.Type)
	a.Equal("/foo?q=bar", data.URL)
	a.Equal("user-id-xyz", data.UserCorrelationId)
}
