// Code generated by MockGen. DO NOT EDIT.
// Source: message.go
//
// Generated by this command:
//
//	mockgen -destination=../mocks/message/mock_message.go -package=message -source=message.go
//

// Package message is a generated GoMock package.
package message

import (
	context "context"
	reflect "reflect"

	config "github.com/piusalfred/whatsapp/config"
	message "github.com/piusalfred/whatsapp/message"
	gomock "go.uber.org/mock/gomock"
)

// MockService is a mock of Service interface.
type MockService struct {
	ctrl     *gomock.Controller
	recorder *MockServiceMockRecorder
	isgomock struct{}
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

// RequestLocation mocks base method.
func (m *MockService) RequestLocation(ctx context.Context, request *message.Request[string]) (*message.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RequestLocation", ctx, request)
	ret0, _ := ret[0].(*message.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RequestLocation indicates an expected call of RequestLocation.
func (mr *MockServiceMockRecorder) RequestLocation(ctx, request any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RequestLocation", reflect.TypeOf((*MockService)(nil).RequestLocation), ctx, request)
}

// SendAudio mocks base method.
func (m *MockService) SendAudio(ctx context.Context, request *message.Request[message.Audio]) (*message.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendAudio", ctx, request)
	ret0, _ := ret[0].(*message.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SendAudio indicates an expected call of SendAudio.
func (mr *MockServiceMockRecorder) SendAudio(ctx, request any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendAudio", reflect.TypeOf((*MockService)(nil).SendAudio), ctx, request)
}

// SendContacts mocks base method.
func (m *MockService) SendContacts(ctx context.Context, request *message.Request[message.Contacts]) (*message.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendContacts", ctx, request)
	ret0, _ := ret[0].(*message.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SendContacts indicates an expected call of SendContacts.
func (mr *MockServiceMockRecorder) SendContacts(ctx, request any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendContacts", reflect.TypeOf((*MockService)(nil).SendContacts), ctx, request)
}

// SendDocument mocks base method.
func (m *MockService) SendDocument(ctx context.Context, request *message.Request[message.Document]) (*message.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendDocument", ctx, request)
	ret0, _ := ret[0].(*message.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SendDocument indicates an expected call of SendDocument.
func (mr *MockServiceMockRecorder) SendDocument(ctx, request any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendDocument", reflect.TypeOf((*MockService)(nil).SendDocument), ctx, request)
}

// SendImage mocks base method.
func (m *MockService) SendImage(ctx context.Context, request *message.Request[message.Image]) (*message.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendImage", ctx, request)
	ret0, _ := ret[0].(*message.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SendImage indicates an expected call of SendImage.
func (mr *MockServiceMockRecorder) SendImage(ctx, request any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendImage", reflect.TypeOf((*MockService)(nil).SendImage), ctx, request)
}

// SendInteractiveMessage mocks base method.
func (m *MockService) SendInteractiveMessage(ctx context.Context, request *message.Request[message.Interactive]) (*message.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendInteractiveMessage", ctx, request)
	ret0, _ := ret[0].(*message.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SendInteractiveMessage indicates an expected call of SendInteractiveMessage.
func (mr *MockServiceMockRecorder) SendInteractiveMessage(ctx, request any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendInteractiveMessage", reflect.TypeOf((*MockService)(nil).SendInteractiveMessage), ctx, request)
}

// SendLocation mocks base method.
func (m *MockService) SendLocation(ctx context.Context, request *message.Request[message.Location]) (*message.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendLocation", ctx, request)
	ret0, _ := ret[0].(*message.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SendLocation indicates an expected call of SendLocation.
func (mr *MockServiceMockRecorder) SendLocation(ctx, request any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendLocation", reflect.TypeOf((*MockService)(nil).SendLocation), ctx, request)
}

// SendReaction mocks base method.
func (m *MockService) SendReaction(ctx context.Context, request *message.Request[message.Reaction]) (*message.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendReaction", ctx, request)
	ret0, _ := ret[0].(*message.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SendReaction indicates an expected call of SendReaction.
func (mr *MockServiceMockRecorder) SendReaction(ctx, request any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendReaction", reflect.TypeOf((*MockService)(nil).SendReaction), ctx, request)
}

// SendSticker mocks base method.
func (m *MockService) SendSticker(ctx context.Context, request *message.Request[message.Sticker]) (*message.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendSticker", ctx, request)
	ret0, _ := ret[0].(*message.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SendSticker indicates an expected call of SendSticker.
func (mr *MockServiceMockRecorder) SendSticker(ctx, request any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendSticker", reflect.TypeOf((*MockService)(nil).SendSticker), ctx, request)
}

// SendTemplate mocks base method.
func (m *MockService) SendTemplate(ctx context.Context, request *message.Request[message.Template]) (*message.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendTemplate", ctx, request)
	ret0, _ := ret[0].(*message.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SendTemplate indicates an expected call of SendTemplate.
func (mr *MockServiceMockRecorder) SendTemplate(ctx, request any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendTemplate", reflect.TypeOf((*MockService)(nil).SendTemplate), ctx, request)
}

// SendText mocks base method.
func (m *MockService) SendText(ctx context.Context, request *message.Request[message.Text]) (*message.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendText", ctx, request)
	ret0, _ := ret[0].(*message.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SendText indicates an expected call of SendText.
func (mr *MockServiceMockRecorder) SendText(ctx, request any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendText", reflect.TypeOf((*MockService)(nil).SendText), ctx, request)
}

// SendVideo mocks base method.
func (m *MockService) SendVideo(ctx context.Context, request *message.Request[message.Video]) (*message.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendVideo", ctx, request)
	ret0, _ := ret[0].(*message.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SendVideo indicates an expected call of SendVideo.
func (mr *MockServiceMockRecorder) SendVideo(ctx, request any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendVideo", reflect.TypeOf((*MockService)(nil).SendVideo), ctx, request)
}

// MockStatusUpdater is a mock of StatusUpdater interface.
type MockStatusUpdater struct {
	ctrl     *gomock.Controller
	recorder *MockStatusUpdaterMockRecorder
	isgomock struct{}
}

// MockStatusUpdaterMockRecorder is the mock recorder for MockStatusUpdater.
type MockStatusUpdaterMockRecorder struct {
	mock *MockStatusUpdater
}

// NewMockStatusUpdater creates a new mock instance.
func NewMockStatusUpdater(ctrl *gomock.Controller) *MockStatusUpdater {
	mock := &MockStatusUpdater{ctrl: ctrl}
	mock.recorder = &MockStatusUpdaterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStatusUpdater) EXPECT() *MockStatusUpdaterMockRecorder {
	return m.recorder
}

// UpdateStatus mocks base method.
func (m *MockStatusUpdater) UpdateStatus(ctx context.Context, request *message.StatusUpdateRequest) (*message.StatusUpdateResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateStatus", ctx, request)
	ret0, _ := ret[0].(*message.StatusUpdateResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateStatus indicates an expected call of UpdateStatus.
func (mr *MockStatusUpdaterMockRecorder) UpdateStatus(ctx, request any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateStatus", reflect.TypeOf((*MockStatusUpdater)(nil).UpdateStatus), ctx, request)
}

// MockSender is a mock of Sender interface.
type MockSender struct {
	ctrl     *gomock.Controller
	recorder *MockSenderMockRecorder
	isgomock struct{}
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
func (m *MockSender) Send(ctx context.Context, conf *config.Config, request *message.BaseRequest) (*message.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Send", ctx, conf, request)
	ret0, _ := ret[0].(*message.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Send indicates an expected call of Send.
func (mr *MockSenderMockRecorder) Send(ctx, conf, request any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockSender)(nil).Send), ctx, conf, request)
}
