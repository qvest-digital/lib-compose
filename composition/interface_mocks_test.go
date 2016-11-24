// Automatically generated by MockGen. DO NOT EDIT!
// Source: github.com/tarent/lib-compose/composition (interfaces: Fragment,ContentLoader,Content,ContentMerger,ContentParser,ResponseProcessor,Cache)

package composition

import (
	gomock "github.com/golang/mock/gomock"
	
	io "io"
	http "net/http"
)

// Mock of Fragment interface
type MockFragment struct {
	ctrl     *gomock.Controller
	recorder *_MockFragmentRecorder
}

// Recorder for MockFragment (not exported)
type _MockFragmentRecorder struct {
	mock *MockFragment
}

func NewMockFragment(ctrl *gomock.Controller) *MockFragment {
	mock := &MockFragment{ctrl: ctrl}
	mock.recorder = &_MockFragmentRecorder{mock}
	return mock
}

func (_m *MockFragment) EXPECT() *_MockFragmentRecorder {
	return _m.recorder
}

func (_m *MockFragment) Execute(_param0 io.Writer, _param1 map[string]interface{}, _param2 func(string) error) error {
	ret := _m.ctrl.Call(_m, "Execute", _param0, _param1, _param2)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockFragmentRecorder) Execute(arg0, arg1, arg2 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Execute", arg0, arg1, arg2)
}

func (_m *MockFragment) MemorySize() int {
	ret := _m.ctrl.Call(_m, "MemorySize")
	ret0, _ := ret[0].(int)
	return ret0
}

func (_mr *_MockFragmentRecorder) MemorySize() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "MemorySize")
}

// Mock of ContentLoader interface
type MockContentLoader struct {
	ctrl     *gomock.Controller
	recorder *_MockContentLoaderRecorder
}

// Recorder for MockContentLoader (not exported)
type _MockContentLoaderRecorder struct {
	mock *MockContentLoader
}

func NewMockContentLoader(ctrl *gomock.Controller) *MockContentLoader {
	mock := &MockContentLoader{ctrl: ctrl}
	mock.recorder = &_MockContentLoaderRecorder{mock}
	return mock
}

func (_m *MockContentLoader) EXPECT() *_MockContentLoaderRecorder {
	return _m.recorder
}

func (_m *MockContentLoader) Load(_param0 *FetchDefinition) (Content, error) {
	ret := _m.ctrl.Call(_m, "Load", _param0)
	ret0, _ := ret[0].(Content)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockContentLoaderRecorder) Load(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Load", arg0)
}

// Mock of Content interface
type MockContent struct {
	ctrl     *gomock.Controller
	recorder *_MockContentRecorder
}

// Recorder for MockContent (not exported)
type _MockContentRecorder struct {
	mock *MockContent
}

func NewMockContent(ctrl *gomock.Controller) *MockContent {
	mock := &MockContent{ctrl: ctrl}
	mock.recorder = &_MockContentRecorder{mock}
	return mock
}

func (_m *MockContent) EXPECT() *_MockContentRecorder {
	return _m.recorder
}

func (_m *MockContent) Body() map[string]Fragment {
	ret := _m.ctrl.Call(_m, "Body")
	ret0, _ := ret[0].(map[string]Fragment)
	return ret0
}

func (_mr *_MockContentRecorder) Body() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Body")
}

func (_m *MockContent) BodyAttributes() Fragment {
	ret := _m.ctrl.Call(_m, "BodyAttributes")
	ret0, _ := ret[0].(Fragment)
	return ret0
}

func (_mr *_MockContentRecorder) BodyAttributes() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "BodyAttributes")
}

func (_m *MockContent) Head() Fragment {
	ret := _m.ctrl.Call(_m, "Head")
	ret0, _ := ret[0].(Fragment)
	return ret0
}

func (_mr *_MockContentRecorder) Head() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Head")
}

func (_m *MockContent) HttpHeader() http.Header {
	ret := _m.ctrl.Call(_m, "HttpHeader")
	ret0, _ := ret[0].(http.Header)
	return ret0
}

func (_mr *_MockContentRecorder) HttpHeader() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "HttpHeader")
}

func (_m *MockContent) HttpStatusCode() int {
	ret := _m.ctrl.Call(_m, "HttpStatusCode")
	ret0, _ := ret[0].(int)
	return ret0
}

func (_mr *_MockContentRecorder) HttpStatusCode() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "HttpStatusCode")
}

func (_m *MockContent) MemorySize() int {
	ret := _m.ctrl.Call(_m, "MemorySize")
	ret0, _ := ret[0].(int)
	return ret0
}

func (_mr *_MockContentRecorder) MemorySize() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "MemorySize")
}

func (_m *MockContent) Meta() map[string]interface{} {
	ret := _m.ctrl.Call(_m, "Meta")
	ret0, _ := ret[0].(map[string]interface{})
	return ret0
}

func (_mr *_MockContentRecorder) Meta() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Meta")
}

func (_m *MockContent) Name() string {
	ret := _m.ctrl.Call(_m, "Name")
	ret0, _ := ret[0].(string)
	return ret0
}

func (_mr *_MockContentRecorder) Name() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Name")
}

func (_m *MockContent) Reader() io.ReadCloser {
	ret := _m.ctrl.Call(_m, "Reader")
	ret0, _ := ret[0].(io.ReadCloser)
	return ret0
}

func (_mr *_MockContentRecorder) Reader() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Reader")
}

func (_m *MockContent) RequiredContent() []*FetchDefinition {
	ret := _m.ctrl.Call(_m, "RequiredContent")
	ret0, _ := ret[0].([]*FetchDefinition)
	return ret0
}

func (_mr *_MockContentRecorder) RequiredContent() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "RequiredContent")
}

func (_m *MockContent) Tail() Fragment {
	ret := _m.ctrl.Call(_m, "Tail")
	ret0, _ := ret[0].(Fragment)
	return ret0
}

func (_mr *_MockContentRecorder) Tail() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Tail")
}

// Mock of ContentMerger interface
type MockContentMerger struct {
	ctrl     *gomock.Controller
	recorder *_MockContentMergerRecorder
}

// Recorder for MockContentMerger (not exported)
type _MockContentMergerRecorder struct {
	mock *MockContentMerger
}

func NewMockContentMerger(ctrl *gomock.Controller) *MockContentMerger {
	mock := &MockContentMerger{ctrl: ctrl}
	mock.recorder = &_MockContentMergerRecorder{mock}
	return mock
}

func (_m *MockContentMerger) EXPECT() *_MockContentMergerRecorder {
	return _m.recorder
}

func (_m *MockContentMerger) AddContent(_param0 Content, _param1 int) {
	_m.ctrl.Call(_m, "AddContent", _param0, _param1)
}

func (_mr *_MockContentMergerRecorder) AddContent(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "AddContent", arg0, arg1)
}

func (_m *MockContentMerger) GetHashes() []string {
	ret := _m.ctrl.Call(_m, "GetHashes")
	ret0, _ := ret[0].([]string)
	return ret0
}

func (_mr *_MockContentMergerRecorder) GetHashes() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetHashes")
}

func (_m *MockContentMerger) GetHtml() ([]byte, error) {
	ret := _m.ctrl.Call(_m, "GetHtml")
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockContentMergerRecorder) GetHtml() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetHtml")
}

// Mock of ContentParser interface
type MockContentParser struct {
	ctrl     *gomock.Controller
	recorder *_MockContentParserRecorder
}

// Recorder for MockContentParser (not exported)
type _MockContentParserRecorder struct {
	mock *MockContentParser
}

func NewMockContentParser(ctrl *gomock.Controller) *MockContentParser {
	mock := &MockContentParser{ctrl: ctrl}
	mock.recorder = &_MockContentParserRecorder{mock}
	return mock
}

func (_m *MockContentParser) EXPECT() *_MockContentParserRecorder {
	return _m.recorder
}

func (_m *MockContentParser) Parse(_param0 *MemoryContent, _param1 io.Reader) error {
	ret := _m.ctrl.Call(_m, "Parse", _param0, _param1)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockContentParserRecorder) Parse(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Parse", arg0, arg1)
}

// Mock of ResponseProcessor interface
type MockResponseProcessor struct {
	ctrl     *gomock.Controller
	recorder *_MockResponseProcessorRecorder
}

// Recorder for MockResponseProcessor (not exported)
type _MockResponseProcessorRecorder struct {
	mock *MockResponseProcessor
}

func NewMockResponseProcessor(ctrl *gomock.Controller) *MockResponseProcessor {
	mock := &MockResponseProcessor{ctrl: ctrl}
	mock.recorder = &_MockResponseProcessorRecorder{mock}
	return mock
}

func (_m *MockResponseProcessor) EXPECT() *_MockResponseProcessorRecorder {
	return _m.recorder
}

func (_m *MockResponseProcessor) Process(_param0 *http.Response, _param1 string) error {
	ret := _m.ctrl.Call(_m, "Process", _param0, _param1)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockResponseProcessorRecorder) Process(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Process", arg0, arg1)
}

// Mock of Cache interface
type MockCache struct {
	ctrl     *gomock.Controller
	recorder *_MockCacheRecorder
}

// Recorder for MockCache (not exported)
type _MockCacheRecorder struct {
	mock *MockCache
}

func NewMockCache(ctrl *gomock.Controller) *MockCache {
	mock := &MockCache{ctrl: ctrl}
	mock.recorder = &_MockCacheRecorder{mock}
	return mock
}

func (_m *MockCache) EXPECT() *_MockCacheRecorder {
	return _m.recorder
}

func (_m *MockCache) Get(_param0 string) (interface{}, bool) {
	ret := _m.ctrl.Call(_m, "Get", _param0)
	ret0, _ := ret[0].(interface{})
	ret1, _ := ret[1].(bool)
	return ret0, ret1
}

func (_mr *_MockCacheRecorder) Get(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Get", arg0)
}

func (_m *MockCache) Invalidate() {
	_m.ctrl.Call(_m, "Invalidate")
}

func (_mr *_MockCacheRecorder) Invalidate() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Invalidate")
}

func (_m *MockCache) PurgeEntries(_param0 []string) {
	_m.ctrl.Call(_m, "PurgeEntries", _param0)
}

func (_mr *_MockCacheRecorder) PurgeEntries(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "PurgeEntries", arg0)
}

func (_m *MockCache) Set(_param0 string, _param1 string, _param2 int, _param3 interface{}) {
	_m.ctrl.Call(_m, "Set", _param0, _param1, _param2, _param3)
}

func (_mr *_MockCacheRecorder) Set(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Set", arg0, arg1, arg2, arg3)
}
