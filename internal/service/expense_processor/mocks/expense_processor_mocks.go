// Code generated by MockGen. DO NOT EDIT.
// Source: internal/service/expense_processor/expense_processor.go

// Package mock_expense_processor is a generated GoMock package.
package mock_expense_processor

import (
	context "context"
	reflect "reflect"
	time "time"

	gomock "github.com/golang/mock/gomock"
	model "gitlab.ozon.dev/cranky4/tg-bot/internal/model"
)

// MockExpenseProcessor is a mock of ExpenseProcessor interface.
type MockExpenseProcessor struct {
	ctrl     *gomock.Controller
	recorder *MockExpenseProcessorMockRecorder
}

// MockExpenseProcessorMockRecorder is the mock recorder for MockExpenseProcessor.
type MockExpenseProcessorMockRecorder struct {
	mock *MockExpenseProcessor
}

// NewMockExpenseProcessor creates a new mock instance.
func NewMockExpenseProcessor(ctrl *gomock.Controller) *MockExpenseProcessor {
	mock := &MockExpenseProcessor{ctrl: ctrl}
	mock.recorder = &MockExpenseProcessorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockExpenseProcessor) EXPECT() *MockExpenseProcessorMockRecorder {
	return m.recorder
}

// AddExpense mocks base method.
func (m *MockExpenseProcessor) AddExpense(ctx context.Context, amount float64, currency, category string, datetime time.Time, userId int64) (*model.Expense, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddExpense", ctx, amount, currency, category, datetime, userId)
	ret0, _ := ret[0].(*model.Expense)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AddExpense indicates an expected call of AddExpense.
func (mr *MockExpenseProcessorMockRecorder) AddExpense(ctx, amount, currency, category, datetime, userId interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddExpense", reflect.TypeOf((*MockExpenseProcessor)(nil).AddExpense), ctx, amount, currency, category, datetime, userId)
}

// GetFreeLimit mocks base method.
func (m *MockExpenseProcessor) GetFreeLimit(ctx context.Context, category, currency string, userId int64) (float64, bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetFreeLimit", ctx, category, currency, userId)
	ret0, _ := ret[0].(float64)
	ret1, _ := ret[1].(bool)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetFreeLimit indicates an expected call of GetFreeLimit.
func (mr *MockExpenseProcessorMockRecorder) GetFreeLimit(ctx, category, currency, userId interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFreeLimit", reflect.TypeOf((*MockExpenseProcessor)(nil).GetFreeLimit), ctx, category, currency, userId)
}

// SetLimit mocks base method.
func (m *MockExpenseProcessor) SetLimit(ctx context.Context, category string, userId int64, amount float64, currency string) (float64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetLimit", ctx, category, userId, amount, currency)
	ret0, _ := ret[0].(float64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SetLimit indicates an expected call of SetLimit.
func (mr *MockExpenseProcessorMockRecorder) SetLimit(ctx, category, userId, amount, currency interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetLimit", reflect.TypeOf((*MockExpenseProcessor)(nil).SetLimit), ctx, category, userId, amount, currency)
}