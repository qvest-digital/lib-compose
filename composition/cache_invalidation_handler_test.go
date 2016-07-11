package composition

import (
	"github.com/golang/mock/gomock"
	mockhttp "github.com/tarent/lib-compose/composition/mocks/net/http"
	"net/http"
	"testing"
)

func Test_CacheInvalidationHandler_Invalidation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	//given
	cacheMocK := NewMockCache(ctrl)
	cih := NewCacheInvalidationHandler(cacheMocK, nil)
	request, _ := http.NewRequest(http.MethodDelete, "internal/cache", nil)

	//when
	cacheMocK.EXPECT().Invalidate().Times(1)
	cih.ServeHTTP(nil, request)
}

func Test_CacheInvalidationHandler_Delegate_Is_Called(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	//given
	handlerMock := mockhttp.NewMockHandler(ctrl)
	cacheMocK := NewMockCache(ctrl)
	cih := NewCacheInvalidationHandler(cacheMocK, handlerMock)
	request, _ := http.NewRequest(http.MethodDelete, "internal/cache", nil)

	//when
	cacheMocK.EXPECT().Invalidate().AnyTimes()
	handlerMock.EXPECT().ServeHTTP(gomock.Any(), gomock.Any()).Times(1)
	cih.ServeHTTP(nil, request)
}
