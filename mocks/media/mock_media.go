// Code generated by MockGen. DO NOT EDIT.
// Source: media.go
//
// Generated by this command:
//
//	mockgen -destination=../mocks/media/mock_media.go -package=media -source=media.go
//

// Package media is a generated GoMock package.
package media

import (
	context "context"
	reflect "reflect"

	media "github.com/piusalfred/whatsapp/media"
	http "github.com/piusalfred/whatsapp/pkg/http"
	gomock "go.uber.org/mock/gomock"
)

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

// Delete mocks base method.
func (m *MockService) Delete(ctx context.Context, request *media.BaseRequest) (*media.DeleteMediaResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", ctx, request)
	ret0, _ := ret[0].(*media.DeleteMediaResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Delete indicates an expected call of Delete.
func (mr *MockServiceMockRecorder) Delete(ctx, request any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockService)(nil).Delete), ctx, request)
}

// Download mocks base method.
func (m *MockService) Download(ctx context.Context, request *media.DownloadRequest, decoder http.ResponseDecoder) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Download", ctx, request, decoder)
	ret0, _ := ret[0].(error)
	return ret0
}

// Download indicates an expected call of Download.
func (mr *MockServiceMockRecorder) Download(ctx, request, decoder any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Download", reflect.TypeOf((*MockService)(nil).Download), ctx, request, decoder)
}

// GetInfo mocks base method.
func (m *MockService) GetInfo(ctx context.Context, request *media.BaseRequest) (*media.Information, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetInfo", ctx, request)
	ret0, _ := ret[0].(*media.Information)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetInfo indicates an expected call of GetInfo.
func (mr *MockServiceMockRecorder) GetInfo(ctx, request any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetInfo", reflect.TypeOf((*MockService)(nil).GetInfo), ctx, request)
}

// Upload mocks base method.
func (m *MockService) Upload(ctx context.Context, req *media.UploadRequest) (*media.UploadMediaResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Upload", ctx, req)
	ret0, _ := ret[0].(*media.UploadMediaResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Upload indicates an expected call of Upload.
func (mr *MockServiceMockRecorder) Upload(ctx, req any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Upload", reflect.TypeOf((*MockService)(nil).Upload), ctx, req)
}