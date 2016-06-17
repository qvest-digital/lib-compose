package logging

import (
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/Sirupsen/logrus/formatters/logstash"
	"net/http"
	"strings"
	"time"
)

var Logger *logrus.Logger

// The of cookies which should not be logged
var AccessLogCookiesBlacklist = []string{}

func init() {
	Set("info", false)
}

// Set creates a new Logger with the matching specification
func Set(level string, textLogging bool) error {
	l, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}

	logger := logrus.New()
	if textLogging {
		logger.Formatter = &logrus.TextFormatter{}
	} else {
		logger.Formatter = &logstash.LogstashFormatter{TimestampFormat: time.RFC3339Nano}
	}
	logger.Level = l
	Logger = logger
	return nil
}

// Access logs an access entry with call duration and status code
func Access(r *http.Request, start time.Time, statusCode int) {
	e := access(r, start, statusCode, nil)

	var msg string
	if len(r.URL.RawQuery) == 0 {
		msg = fmt.Sprintf("%v %v %v", statusCode, r.Method, r.URL.Path)
	} else {
		msg = fmt.Sprintf("%v %v %v?...", statusCode, r.Method, r.URL.Path)
	}

	if statusCode >= 200 && statusCode < 399 {
		e.Info(msg)
	} else if statusCode >= 400 && statusCode < 499 {
		e.Warn(msg)
	} else {
		e.Error(msg)
	}
}

// AccessError logs an error while accessing
func AccessError(r *http.Request, start time.Time, err error) {
	e := access(r, start, 0, err)
	e.Errorf("ERROR %v %v", r.Method, r.URL.Path)
}

func access(r *http.Request, start time.Time, statusCode int, err error) *logrus.Entry {
	url := r.URL.Path
	if r.URL.RawQuery != "" {
		url += "?" + r.URL.RawQuery
	}
	fields := logrus.Fields{
		"type":       "access",
		"@timestamp": start,
		"remote_ip":  getRemoteIp(r),
		"host":       r.Host,
		"url":        url,
		"method":     r.Method,
		"duration":   time.Since(start).Nanoseconds() / 1000000,
	}

	if statusCode != 0 {
		fields["response_status"] = statusCode
	}

	if err != nil {
		fields[logrus.ErrorKey] = err
	}

	correlationId := GetCorrelationId(r)
	if correlationId != "" {
		fields["correlation_id"] = correlationId
	}

	userCorrelationId := GetUserCorrelationId(r)
	if userCorrelationId != "" {
		fields["user_correlation_id"] = userCorrelationId
	}

	cookies := map[string]string{}
	for _, c := range r.Cookies() {
		if !contains(AccessLogCookiesBlacklist, c.Name) {
			cookies[c.Name] = c.Value
		}
	}
	if len(cookies) > 0 {
		fields["cookies"] = cookies
	}

	return Logger.WithFields(fields)
}

// ApplicationStart logs the start of an application
// with the configuration struct or map as paramter.
func ApplicationStart(appName string, args interface{}) {
	fields := logrus.Fields{}

	jsonString, err := json.Marshal(args)
	if err == nil {
		err := json.Unmarshal(jsonString, &fields)
		if err != nil {
			fields["parse_error"] = err.Error()
		}
	}
	fields["type"] = "service-start"

	Logger.WithFields(fields).Infof("starting application: %v", appName)
}

func getRemoteIp(r *http.Request) string {
	if r.Header.Get("X-Cluster-Client-Ip") != "" {
		return r.Header.Get("X-Cluster-Client-Ip")
	}
	if r.Header.Get("X-Real-Ip") != "" {
		return r.Header.Get("X-Real-Ip")
	}
	return strings.Split(r.RemoteAddr, ":")[0]
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
