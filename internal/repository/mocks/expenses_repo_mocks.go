// Code generated by MockGen. DO NOT EDIT.
// Source: internal/repository/expenses.go

// Package mock_repository is a generated GoMock package.
package mock_repository

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	model "gitlab.ozon.dev/cranky4/tg-bot/internal/model"
)

// MockExpensesRepository is a mock of ExpensesRepository interface.
type MockExpensesRepository struct {
	ctrl     *gomock.Controller
	recorder *MockExpensesRepositoryMockRecorder
}

// MockExpensesRepositoryMockRecorder is the mock recorder for MockExpensesRepository.
type MockExpensesRepositoryMockRecorder struct {
	mock *MockExpensesRepository
}

// NewMockExpensesRepository creates a new mock instance.
func NewMockExpensesRepository(ctrl *gomock.Controller) *MockExpensesRepository {
	mock := &MockExpensesRepository{ctrl: ctrl}
	mock.recorder = &MockExpensesRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockExpensesRepository) EXPECT() *MockExpensesRepositoryMockRecorder {
	return m.recorder
}

// Add mocks base method.
func (m *MockExpensesRepository) Add(ctx context.Context, expense model.Expense) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Add", ctx, expense)
	ret0, _ := ret[0].(error)
	return ret0
}

// Add indicates an expected call of Add.
func (mr *MockExpensesRepositoryMockRecorder) Add(ctx, expense interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Add", reflect.TypeOf((*MockExpensesRepository)(nil).Add), ctx, expense)
}

// GetExpenses mocks base method.
func (m *MockExpensesRepository) GetExpenses(ctx context.Context, period model.ExpensePeriod, userId int64) ([]*model.Expense, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetExpenses", ctx, period, userId)
	ret0, _ := ret[0].([]*model.Expense)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetExpenses indicates an expected call of GetExpenses.
func (mr *MockExpensesRepositoryMockRecorder) GetExpenses(ctx, period, userId interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetExpenses", reflect.TypeOf((*MockExpensesRepository)(nil).GetExpenses), ctx, period, userId)
}

// GetFreeLimit mocks base method.
func (m *MockExpensesRepository) GetFreeLimit(ctx context.Context, category string, userId int64) (int64, bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetFreeLimit", ctx, category, userId)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(bool)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetFreeLimit indicates an expected call of GetFreeLimit.
func (mr *MockExpensesRepositoryMockRecorder) GetFreeLimit(ctx, category, userId interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFreeLimit", reflect.TypeOf((*MockExpensesRepository)(nil).GetFreeLimit), ctx, category, userId)
}

// SetLimit mocks base method.
func (m *MockExpensesRepository) SetLimit(ctx context.Context, category string, userId, amount int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetLimit", ctx, category, userId, amount)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetLimit indicates an expected call of SetLimit.
func (mr *MockExpensesRepositoryMockRecorder) SetLimit(ctx, category, userId, amount interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetLimit", reflect.TypeOf((*MockExpensesRepository)(nil).SetLimit), ctx, category, userId, amount)
}
