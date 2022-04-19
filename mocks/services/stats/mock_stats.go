// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/rudderlabs/rudder-server/services/stats (interfaces: Stats,RudderStats)

// Package mock_stats is a generated GoMock package.
package mock_stats

import (
	reflect "reflect"
	time "time"

	gomock "github.com/golang/mock/gomock"
	stats "github.com/rudderlabs/rudder-server/services/stats"
)

// MockStats is a mock of Stats interface.
type MockStats struct {
	ctrl     *gomock.Controller
	recorder *MockStatsMockRecorder
}

// MockStatsMockRecorder is the mock recorder for MockStats.
type MockStatsMockRecorder struct {
	mock *MockStats
}

// NewMockStats creates a new mock instance.
func NewMockStats(ctrl *gomock.Controller) *MockStats {
	mock := &MockStats{ctrl: ctrl}
	mock.recorder = &MockStatsMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStats) EXPECT() *MockStatsMockRecorder {
	return m.recorder
}

// NewSampledTaggedStat mocks base method.
func (m *MockStats) NewSampledTaggedStat(arg0, arg1 string, arg2 stats.Tags) stats.RudderStats {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewSampledTaggedStat", arg0, arg1, arg2)
	ret0, _ := ret[0].(stats.RudderStats)
	return ret0
}

// NewSampledTaggedStat indicates an expected call of NewSampledTaggedStat.
func (mr *MockStatsMockRecorder) NewSampledTaggedStat(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewSampledTaggedStat", reflect.TypeOf((*MockStats)(nil).NewSampledTaggedStat), arg0, arg1, arg2)
}

// NewStat mocks base method.
func (m *MockStats) NewStat(arg0, arg1 string) stats.RudderStats {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewStat", arg0, arg1)
	ret0, _ := ret[0].(stats.RudderStats)
	return ret0
}

// NewStat indicates an expected call of NewStat.
func (mr *MockStatsMockRecorder) NewStat(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewStat", reflect.TypeOf((*MockStats)(nil).NewStat), arg0, arg1)
}

// NewTaggedStat mocks base method.
func (m *MockStats) NewTaggedStat(arg0, arg1 string, arg2 stats.Tags) stats.RudderStats {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewTaggedStat", arg0, arg1, arg2)
	ret0, _ := ret[0].(stats.RudderStats)
	return ret0
}

// NewTaggedStat indicates an expected call of NewTaggedStat.
func (mr *MockStatsMockRecorder) NewTaggedStat(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewTaggedStat", reflect.TypeOf((*MockStats)(nil).NewTaggedStat), arg0, arg1, arg2)
}

// MockRudderStats is a mock of RudderStats interface.
type MockRudderStats struct {
	ctrl     *gomock.Controller
	recorder *MockRudderStatsMockRecorder
}

// MockRudderStatsMockRecorder is the mock recorder for MockRudderStats.
type MockRudderStatsMockRecorder struct {
	mock *MockRudderStats
}

// NewMockRudderStats creates a new mock instance.
func NewMockRudderStats(ctrl *gomock.Controller) *MockRudderStats {
	mock := &MockRudderStats{ctrl: ctrl}
	mock.recorder = &MockRudderStatsMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRudderStats) EXPECT() *MockRudderStatsMockRecorder {
	return m.recorder
}

// Count mocks base method.
func (m *MockRudderStats) Count(arg0 int) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Count", arg0)
}

// Count indicates an expected call of Count.
func (mr *MockRudderStatsMockRecorder) Count(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Count", reflect.TypeOf((*MockRudderStats)(nil).Count), arg0)
}

// DeferredTimer mocks base method.
func (m *MockRudderStats) DeferredTimer() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "DeferredTimer")
}

// DeferredTimer indicates an expected call of DeferredTimer.
func (mr *MockRudderStatsMockRecorder) DeferredTimer() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeferredTimer", reflect.TypeOf((*MockRudderStats)(nil).DeferredTimer))
}

// End mocks base method.
func (m *MockRudderStats) End() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "End")
}

// End indicates an expected call of End.
func (mr *MockRudderStatsMockRecorder) End() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "End", reflect.TypeOf((*MockRudderStats)(nil).End))
}

// Gauge mocks base method.
func (m *MockRudderStats) Gauge(arg0 interface{}) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Gauge", arg0)
}

// Gauge indicates an expected call of Gauge.
func (mr *MockRudderStatsMockRecorder) Gauge(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Gauge", reflect.TypeOf((*MockRudderStats)(nil).Gauge), arg0)
}

// Increment mocks base method.
func (m *MockRudderStats) Increment() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Increment")
}

// Increment indicates an expected call of Increment.
func (mr *MockRudderStatsMockRecorder) Increment() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Increment", reflect.TypeOf((*MockRudderStats)(nil).Increment))
}

// Observe mocks base method.
func (m *MockRudderStats) Observe(arg0 float64) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Observe", arg0)
}

// Observe indicates an expected call of Observe.
func (mr *MockRudderStatsMockRecorder) Observe(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Observe", reflect.TypeOf((*MockRudderStats)(nil).Observe), arg0)
}

// SendTiming mocks base method.
func (m *MockRudderStats) SendTiming(arg0 time.Duration) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SendTiming", arg0)
}

// SendTiming indicates an expected call of SendTiming.
func (mr *MockRudderStatsMockRecorder) SendTiming(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendTiming", reflect.TypeOf((*MockRudderStats)(nil).SendTiming), arg0)
}

// Since mocks base method.
func (m *MockRudderStats) Since(arg0 time.Time) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Since", arg0)
}

// Since indicates an expected call of Since.
func (mr *MockRudderStatsMockRecorder) Since(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Since", reflect.TypeOf((*MockRudderStats)(nil).Since), arg0)
}

// Start mocks base method.
func (m *MockRudderStats) Start() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Start")
}

// Start indicates an expected call of Start.
func (mr *MockRudderStatsMockRecorder) Start() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockRudderStats)(nil).Start))
}
