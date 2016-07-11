package composition

import (
	"content-ui-service/mocks/net/http"
	"github.com/golang/mock/gomock"
	"testing"
)

func Test_CacheInvalidationHandler_Invalidation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	//given
	cacheMocK := NewMockCache(ctrl)
	cih := &CacheInvalidationHandler{cache: cacheMocK}

	//when
	cacheMocK.EXPECT().Invalidate().Times(1)
	cih.ServeHTTP(nil, nil)
}

func Test_CacheInvalidationHandler_Delegate_Is_Called(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	//given
	handlerMock := http.NewMockHandler(ctrl)
	cacheMocK := NewMockCache(ctrl)
	cih := &CacheInvalidationHandler{cache: cacheMocK, next: handlerMock}

	//when
	cacheMocK.EXPECT().Invalidate().AnyTimes()
	handlerMock.EXPECT().ServeHTTP(gomock.Any(), gomock.Any()).Times(1)
	cih.ServeHTTP(nil, nil)
}
