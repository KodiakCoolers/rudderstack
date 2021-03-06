// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/rudderlabs/rudder-server/utils/pubsub (interfaces: PublishSubscriber)

// Package utils is a generated GoMock package.
package utils

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	pubsub "github.com/rudderlabs/rudder-server/utils/pubsub"
)

// MockPublishSubscriber is a mock of PublishSubscriber interface.
type MockPublishSubscriber struct {
	ctrl     *gomock.Controller
	recorder *MockPublishSubscriberMockRecorder
}

// MockPublishSubscriberMockRecorder is the mock recorder for MockPublishSubscriber.
type MockPublishSubscriberMockRecorder struct {
	mock *MockPublishSubscriber
}

// NewMockPublishSubscriber creates a new mock instance.
func NewMockPublishSubscriber(ctrl *gomock.Controller) *MockPublishSubscriber {
	mock := &MockPublishSubscriber{ctrl: ctrl}
	mock.recorder = &MockPublishSubscriberMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPublishSubscriber) EXPECT() *MockPublishSubscriberMockRecorder {
	return m.recorder
}

// Publish mocks base method.
func (m *MockPublishSubscriber) Publish(arg0 string, arg1 interface{}) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Publish", arg0, arg1)
}

// Publish indicates an expected call of Publish.
func (mr *MockPublishSubscriberMockRecorder) Publish(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Publish", reflect.TypeOf((*MockPublishSubscriber)(nil).Publish), arg0, arg1)
}

// Subscribe mocks base method.
func (m *MockPublishSubscriber) Subscribe(arg0 string, arg1 pubsub.DataChannel) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Subscribe", arg0, arg1)
}

// Subscribe indicates an expected call of Subscribe.
func (mr *MockPublishSubscriberMockRecorder) Subscribe(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Subscribe", reflect.TypeOf((*MockPublishSubscriber)(nil).Subscribe), arg0, arg1)
}
