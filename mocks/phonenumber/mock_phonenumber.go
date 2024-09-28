// Code generated by MockGen. DO NOT EDIT.
// Source: phonenumber.go
//
// Generated by this command:
//
//	mockgen -destination=../mocks/phonenumber/mock_phonenumber.go -package=phonenumber -source=phonenumber.go
//

// Package phonenumber is a generated GoMock package.
package phonenumber

import (
	context "context"
	reflect "reflect"

	config "github.com/piusalfred/whatsapp/config"
	phonenumber "github.com/piusalfred/whatsapp/phonenumber"
	gomock "go.uber.org/mock/gomock"
)

// MockSender is a mock of Sender interface.
type MockSender struct {
	ctrl     *gomock.Controller
	recorder *MockSenderMockRecorder
}

// MockSenderMockRecorder is the mock recorder for MockSender.
type MockSenderMockRecorder struct {
	mock *MockSender
}

// NewMockSender creates a new mock instance.
func NewMockSender(ctrl *gomock.Controller) *MockSender {
	mock := &MockSender{ctrl: ctrl}
	mock.recorder = &MockSenderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSender) EXPECT() *MockSenderMockRecorder {
	return m.recorder
}

// Send mocks base method.
func (m *MockSender) Send(ctx context.Context, conf *config.Config, req *phonenumber.BaseRequest) (*phonenumber.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Send", ctx, conf, req)
	ret0, _ := ret[0].(*phonenumber.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Send indicates an expected call of Send.
func (mr *MockSenderMockRecorder) Send(ctx, conf, req any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockSender)(nil).Send), ctx, conf, req)
}

// MockService is a mock of Service interface.
type MockService struct {
	ctrl     *gomock.Controller
	recorder *MockServiceMockRecorder
}

// MockServiceMockRecorder is the mock recorder for MockService.
type MockServiceMockRecorder struct {
	mock *MockService
}

// NewMockService creates a new mock instance.
func NewMockService(ctrl *gomock.Controller) *MockService {
	mock := &MockService{ctrl: ctrl}
	mock.recorder = &MockServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockService) EXPECT() *MockServiceMockRecorder {
	return m.recorder
}

// Get mocks base method.
func (m *MockService) Get(ctx context.Context, request *phonenumber.GetRequest) (*phonenumber.PhoneNumber, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx, request)
	ret0, _ := ret[0].(*phonenumber.PhoneNumber)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockServiceMockRecorder) Get(ctx, request any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockService)(nil).Get), ctx, request)
}

// List mocks base method.
func (m *MockService) List(ctx context.Context) (*phonenumber.ListResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", ctx)
	ret0, _ := ret[0].(*phonenumber.ListResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List.
func (mr *MockServiceMockRecorder) List(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockService)(nil).List), ctx)
}
