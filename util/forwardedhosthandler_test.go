package util

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	mockHttp "lib-compose/composition/mocks/net/http"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_DelegateIsCalled(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	mockDelegator := mockHttp.NewMockHandler(ctl)
	fhh := NewForwardedHostHandler(mockDelegator)

	req, _ := http.NewRequest("GET", "", nil)
	resp := httptest.NewRecorder()

	//expected the delegate should not been called
	mockDelegator.EXPECT().ServeHTTP(gomock.Any(), gomock.Any()).Times(1)

	// when
	fhh.ServeHTTP(resp, req)
}

func Test_XForwardeHostHeaderIsSet(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	fhh := ForwardedHostHandler{}

	req, _ := http.NewRequest("GET", "", nil)

	expected := "testitest"
	req.Host = expected
	resp := httptest.NewRecorder()

	//when
	fhh.ServeHTTP(resp, req)
	assert.Contains(t, req.Header.Get(X_FORWARDED_HOST_HEADER_KEY), expected)
}

func Test_OldXForwardedHostHeaderIsDropped(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	fhh := ForwardedHostHandler{}

	req, _ := http.NewRequest("GET", "", nil)

	notExpected := "nottestitest"
	req.Header.Set(X_FORWARDED_HOST_HEADER_KEY, notExpected)
	resp := httptest.NewRecorder()

	//when
	fhh.ServeHTTP(resp, req)
	assert.Equal(t, req.Header.Get(X_FORWARDED_HOST_HEADER_KEY), "")
}
