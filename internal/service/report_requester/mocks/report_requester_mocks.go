// Code generated by MockGen. DO NOT EDIT.
// Source: internal/service/report_requester/report_requester.go

// Package mock_reportrequester is a generated GoMock package.
package mock_reportrequester

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	model "gitlab.ozon.dev/cranky4/tg-bot/internal/model"
)

// MockReportRequester is a mock of ReportRequester interface.
type MockReportRequester struct {
	ctrl     *gomock.Controller
	recorder *MockReportRequesterMockRecorder
}

// MockReportRequesterMockRecorder is the mock recorder for MockReportRequester.
type MockReportRequesterMockRecorder struct {
	mock *MockReportRequester
}

// NewMockReportRequester creates a new mock instance.
func NewMockReportRequester(ctrl *gomock.Controller) *MockReportRequester {
	mock := &MockReportRequester{ctrl: ctrl}
	mock.recorder = &MockReportRequesterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockReportRequester) EXPECT() *MockReportRequesterMockRecorder {
	return m.recorder
}

// SendRequestReport mocks base method.
func (m *MockReportRequester) SendRequestReport(ctx context.Context, userID int64, period model.ExpensePeriod, currency string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendRequestReport", ctx, userID, period, currency)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendRequestReport indicates an expected call of SendRequestReport.
func (mr *MockReportRequesterMockRecorder) SendRequestReport(ctx, userID, period, currency interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendRequestReport", reflect.TypeOf((*MockReportRequester)(nil).SendRequestReport), ctx, userID, period, currency)
}