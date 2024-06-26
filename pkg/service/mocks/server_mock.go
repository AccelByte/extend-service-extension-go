// Code generated by MockGen. DO NOT EDIT.
// Source: extend-custom-guild-service/pkg/pb (interfaces: MyServiceServer)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	guild "extend-custom-guild-service/pkg/pb"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockMyServiceServer is a mock of MyServiceServer interface.
type MockMyServiceServer struct {
	ctrl     *gomock.Controller
	recorder *MockMyServiceServerMockRecorder
}

// MockMyServiceServerMockRecorder is the mock recorder for MockMyServiceServer.
type MockMyServiceServerMockRecorder struct {
	mock *MockMyServiceServer
}

// NewMockMyServiceServer creates a new mock instance.
func NewMockMyServiceServer(ctrl *gomock.Controller) *MockMyServiceServer {
	mock := &MockMyServiceServer{ctrl: ctrl}
	mock.recorder = &MockMyServiceServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMyServiceServer) EXPECT() *MockMyServiceServerMockRecorder {
	return m.recorder
}

// CreateOrUpdateGuildProgress mocks base method.
func (m *MockMyServiceServer) CreateOrUpdateGuildProgress(arg0 context.Context, arg1 *guild.CreateOrUpdateGuildProgressRequest) (*guild.CreateOrUpdateGuildProgressResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateOrUpdateGuildProgress", arg0, arg1)
	ret0, _ := ret[0].(*guild.CreateOrUpdateGuildProgressResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateOrUpdateGuildProgress indicates an expected call of CreateOrUpdateGuildProgress.
func (mr *MockMyServiceServerMockRecorder) CreateOrUpdateGuildProgress(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateOrUpdateGuildProgress", reflect.TypeOf((*MockMyServiceServer)(nil).CreateOrUpdateGuildProgress), arg0, arg1)
}

// GetGuildProgress mocks base method.
func (m *MockMyServiceServer) GetGuildProgress(arg0 context.Context, arg1 *guild.GetGuildProgressRequest) (*guild.GetGuildProgressResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetGuildProgress", arg0, arg1)
	ret0, _ := ret[0].(*guild.GetGuildProgressResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetGuildProgress indicates an expected call of GetGuildProgress.
func (mr *MockMyServiceServerMockRecorder) GetGuildProgress(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetGuildProgress", reflect.TypeOf((*MockMyServiceServer)(nil).GetGuildProgress), arg0, arg1)
}

// mustEmbedUnimplementedMyServiceServer mocks base method.
func (m *MockMyServiceServer) mustEmbedUnimplementedMyServiceServer() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "mustEmbedUnimplementedMyServiceServer")
}

// mustEmbedUnimplementedMyServiceServer indicates an expected call of mustEmbedUnimplementedMyServiceServer.
func (mr *MockMyServiceServerMockRecorder) mustEmbedUnimplementedMyServiceServer() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "mustEmbedUnimplementedMyServiceServer", reflect.TypeOf((*MockMyServiceServer)(nil).mustEmbedUnimplementedMyServiceServer))
}
