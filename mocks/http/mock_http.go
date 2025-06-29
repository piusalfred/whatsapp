// Code generated by MockGen. DO NOT EDIT.
// Source: ./pkg/http/http.go
//
// Generated by this command:
//
//	mockgen -destination=./mocks/http/mock_http.go -package=http -source=./pkg/http/http.go
//

// Package http is a generated GoMock package.
package http

import (
	context "context"
	http0 "net/http"
	reflect "reflect"

	http "github.com/piusalfred/whatsapp/pkg/http"
	gomock "go.uber.org/mock/gomock"
)

// MockSender is a mock of Sender interface.
type MockSender[T any] struct {
	ctrl     *gomock.Controller
	recorder *MockSenderMockRecorder[T]
	isgomock struct{}
}

// MockSenderMockRecorder is the mock recorder for MockSender.
type MockSenderMockRecorder[T any] struct {
	mock *MockSender[T]
}

// NewMockSender creates a new mock instance.
func NewMockSender[T any](ctrl *gomock.Controller) *MockSender[T] {
	mock := &MockSender[T]{ctrl: ctrl}
	mock.recorder = &MockSenderMockRecorder[T]{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSender[T]) EXPECT() *MockSenderMockRecorder[T] {
	return m.recorder
}

// Send mocks base method.
func (m *MockSender[T]) Send(ctx context.Context, request *http.Request[T], decoder http.ResponseDecoder) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Send", ctx, request, decoder)
	ret0, _ := ret[0].(error)
	return ret0
}

// Send indicates an expected call of Send.
func (mr *MockSenderMockRecorder[T]) Send(ctx, request, decoder any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockSender[T])(nil).Send), ctx, request, decoder)
}

// MockRequestInterceptor is a mock of RequestInterceptor interface.
type MockRequestInterceptor struct {
	ctrl     *gomock.Controller
	recorder *MockRequestInterceptorMockRecorder
	isgomock struct{}
}

// MockRequestInterceptorMockRecorder is the mock recorder for MockRequestInterceptor.
type MockRequestInterceptorMockRecorder struct {
	mock *MockRequestInterceptor
}

// NewMockRequestInterceptor creates a new mock instance.
func NewMockRequestInterceptor(ctrl *gomock.Controller) *MockRequestInterceptor {
	mock := &MockRequestInterceptor{ctrl: ctrl}
	mock.recorder = &MockRequestInterceptorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRequestInterceptor) EXPECT() *MockRequestInterceptorMockRecorder {
	return m.recorder
}

// InterceptRequest mocks base method.
func (m *MockRequestInterceptor) InterceptRequest(ctx context.Context, request *http0.Request) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InterceptRequest", ctx, request)
	ret0, _ := ret[0].(error)
	return ret0
}

// InterceptRequest indicates an expected call of InterceptRequest.
func (mr *MockRequestInterceptorMockRecorder) InterceptRequest(ctx, request any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InterceptRequest", reflect.TypeOf((*MockRequestInterceptor)(nil).InterceptRequest), ctx, request)
}

// MockResponseInterceptor is a mock of ResponseInterceptor interface.
type MockResponseInterceptor struct {
	ctrl     *gomock.Controller
	recorder *MockResponseInterceptorMockRecorder
	isgomock struct{}
}

// MockResponseInterceptorMockRecorder is the mock recorder for MockResponseInterceptor.
type MockResponseInterceptorMockRecorder struct {
	mock *MockResponseInterceptor
}

// NewMockResponseInterceptor creates a new mock instance.
func NewMockResponseInterceptor(ctrl *gomock.Controller) *MockResponseInterceptor {
	mock := &MockResponseInterceptor{ctrl: ctrl}
	mock.recorder = &MockResponseInterceptorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockResponseInterceptor) EXPECT() *MockResponseInterceptorMockRecorder {
	return m.recorder
}

// InterceptResponse mocks base method.
func (m *MockResponseInterceptor) InterceptResponse(ctx context.Context, response *http0.Response) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InterceptResponse", ctx, response)
	ret0, _ := ret[0].(error)
	return ret0
}

// InterceptResponse indicates an expected call of InterceptResponse.
func (mr *MockResponseInterceptorMockRecorder) InterceptResponse(ctx, response any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InterceptResponse", reflect.TypeOf((*MockResponseInterceptor)(nil).InterceptResponse), ctx, response)
}

// MockResponseDecoder is a mock of ResponseDecoder interface.
type MockResponseDecoder struct {
	ctrl     *gomock.Controller
	recorder *MockResponseDecoderMockRecorder
	isgomock struct{}
}

// MockResponseDecoderMockRecorder is the mock recorder for MockResponseDecoder.
type MockResponseDecoderMockRecorder struct {
	mock *MockResponseDecoder
}

// NewMockResponseDecoder creates a new mock instance.
func NewMockResponseDecoder(ctrl *gomock.Controller) *MockResponseDecoder {
	mock := &MockResponseDecoder{ctrl: ctrl}
	mock.recorder = &MockResponseDecoderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockResponseDecoder) EXPECT() *MockResponseDecoderMockRecorder {
	return m.recorder
}

// Decode mocks base method.
func (m *MockResponseDecoder) Decode(ctx context.Context, response *http0.Response) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Decode", ctx, response)
	ret0, _ := ret[0].(error)
	return ret0
}

// Decode indicates an expected call of Decode.
func (mr *MockResponseDecoderMockRecorder) Decode(ctx, response any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Decode", reflect.TypeOf((*MockResponseDecoder)(nil).Decode), ctx, response)
}
