// Code generated by MockGen. DO NOT EDIT.
// Source: gateway.go

// Package mock_gateway is a generated GoMock package.
package mock_gateway

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	v1 "github.com/utilitywarehouse/account-platform-protobuf-model/gen/go/account/api/v1"
	lowribeckv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/lowribeck/v1"
	grpc "google.golang.org/grpc"
)

// MockMachineAuthInjector is a mock of MachineAuthInjector interface.
type MockMachineAuthInjector struct {
	ctrl     *gomock.Controller
	recorder *MockMachineAuthInjectorMockRecorder
}

// MockMachineAuthInjectorMockRecorder is the mock recorder for MockMachineAuthInjector.
type MockMachineAuthInjectorMockRecorder struct {
	mock *MockMachineAuthInjector
}

// NewMockMachineAuthInjector creates a new mock instance.
func NewMockMachineAuthInjector(ctrl *gomock.Controller) *MockMachineAuthInjector {
	mock := &MockMachineAuthInjector{ctrl: ctrl}
	mock.recorder = &MockMachineAuthInjectorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMachineAuthInjector) EXPECT() *MockMachineAuthInjectorMockRecorder {
	return m.recorder
}

// ToCtx mocks base method.
func (m *MockMachineAuthInjector) ToCtx(arg0 context.Context) context.Context {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ToCtx", arg0)
	ret0, _ := ret[0].(context.Context)
	return ret0
}

// ToCtx indicates an expected call of ToCtx.
func (mr *MockMachineAuthInjectorMockRecorder) ToCtx(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ToCtx", reflect.TypeOf((*MockMachineAuthInjector)(nil).ToCtx), arg0)
}

// MockAccountClient is a mock of AccountClient interface.
type MockAccountClient struct {
	ctrl     *gomock.Controller
	recorder *MockAccountClientMockRecorder
}

// MockAccountClientMockRecorder is the mock recorder for MockAccountClient.
type MockAccountClientMockRecorder struct {
	mock *MockAccountClient
}

// NewMockAccountClient creates a new mock instance.
func NewMockAccountClient(ctrl *gomock.Controller) *MockAccountClient {
	mock := &MockAccountClient{ctrl: ctrl}
	mock.recorder = &MockAccountClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAccountClient) EXPECT() *MockAccountClientMockRecorder {
	return m.recorder
}

// GetAccount mocks base method.
func (m *MockAccountClient) GetAccount(ctx context.Context, in *v1.GetAccountRequest, opts ...grpc.CallOption) (*v1.GetAccountResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetAccount", varargs...)
	ret0, _ := ret[0].(*v1.GetAccountResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAccount indicates an expected call of GetAccount.
func (mr *MockAccountClientMockRecorder) GetAccount(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAccount", reflect.TypeOf((*MockAccountClient)(nil).GetAccount), varargs...)
}

// MockAccountNumberClient is a mock of AccountNumberClient interface.
type MockAccountNumberClient struct {
	ctrl     *gomock.Controller
	recorder *MockAccountNumberClientMockRecorder
}

// MockAccountNumberClientMockRecorder is the mock recorder for MockAccountNumberClient.
type MockAccountNumberClientMockRecorder struct {
	mock *MockAccountNumberClient
}

// NewMockAccountNumberClient creates a new mock instance.
func NewMockAccountNumberClient(ctrl *gomock.Controller) *MockAccountNumberClient {
	mock := &MockAccountNumberClient{ctrl: ctrl}
	mock.recorder = &MockAccountNumberClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAccountNumberClient) EXPECT() *MockAccountNumberClientMockRecorder {
	return m.recorder
}

// AccountNumber mocks base method.
func (m *MockAccountNumberClient) AccountNumber(ctx context.Context, in *v1.AccountNumberRequest, opts ...grpc.CallOption) (*v1.AccountNumberResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "AccountNumber", varargs...)
	ret0, _ := ret[0].(*v1.AccountNumberResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AccountNumber indicates an expected call of AccountNumber.
func (mr *MockAccountNumberClientMockRecorder) AccountNumber(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AccountNumber", reflect.TypeOf((*MockAccountNumberClient)(nil).AccountNumber), varargs...)
}

// MockLowriBeckClient is a mock of LowriBeckClient interface.
type MockLowriBeckClient struct {
	ctrl     *gomock.Controller
	recorder *MockLowriBeckClientMockRecorder
}

// MockLowriBeckClientMockRecorder is the mock recorder for MockLowriBeckClient.
type MockLowriBeckClientMockRecorder struct {
	mock *MockLowriBeckClient
}

// NewMockLowriBeckClient creates a new mock instance.
func NewMockLowriBeckClient(ctrl *gomock.Controller) *MockLowriBeckClient {
	mock := &MockLowriBeckClient{ctrl: ctrl}
	mock.recorder = &MockLowriBeckClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockLowriBeckClient) EXPECT() *MockLowriBeckClientMockRecorder {
	return m.recorder
}

// CreateBooking mocks base method.
func (m *MockLowriBeckClient) CreateBooking(ctx context.Context, in *lowribeckv1.CreateBookingRequest, opts ...grpc.CallOption) (*lowribeckv1.CreateBookingResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "CreateBooking", varargs...)
	ret0, _ := ret[0].(*lowribeckv1.CreateBookingResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateBooking indicates an expected call of CreateBooking.
func (mr *MockLowriBeckClientMockRecorder) CreateBooking(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateBooking", reflect.TypeOf((*MockLowriBeckClient)(nil).CreateBooking), varargs...)
}

// CreateBookingPointOfSale mocks base method.
func (m *MockLowriBeckClient) CreateBookingPointOfSale(ctx context.Context, in *lowribeckv1.CreateBookingPointOfSaleRequest, opts ...grpc.CallOption) (*lowribeckv1.CreateBookingPointOfSaleResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "CreateBookingPointOfSale", varargs...)
	ret0, _ := ret[0].(*lowribeckv1.CreateBookingPointOfSaleResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateBookingPointOfSale indicates an expected call of CreateBookingPointOfSale.
func (mr *MockLowriBeckClientMockRecorder) CreateBookingPointOfSale(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateBookingPointOfSale", reflect.TypeOf((*MockLowriBeckClient)(nil).CreateBookingPointOfSale), varargs...)
}

// GetAvailableSlots mocks base method.
func (m *MockLowriBeckClient) GetAvailableSlots(ctx context.Context, in *lowribeckv1.GetAvailableSlotsRequest, opts ...grpc.CallOption) (*lowribeckv1.GetAvailableSlotsResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetAvailableSlots", varargs...)
	ret0, _ := ret[0].(*lowribeckv1.GetAvailableSlotsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAvailableSlots indicates an expected call of GetAvailableSlots.
func (mr *MockLowriBeckClientMockRecorder) GetAvailableSlots(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAvailableSlots", reflect.TypeOf((*MockLowriBeckClient)(nil).GetAvailableSlots), varargs...)
}

// GetAvailableSlotsPointOfSale mocks base method.
func (m *MockLowriBeckClient) GetAvailableSlotsPointOfSale(ctx context.Context, in *lowribeckv1.GetAvailableSlotsPointOfSaleRequest, opts ...grpc.CallOption) (*lowribeckv1.GetAvailableSlotsPointOfSaleResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetAvailableSlotsPointOfSale", varargs...)
	ret0, _ := ret[0].(*lowribeckv1.GetAvailableSlotsPointOfSaleResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAvailableSlotsPointOfSale indicates an expected call of GetAvailableSlotsPointOfSale.
func (mr *MockLowriBeckClientMockRecorder) GetAvailableSlotsPointOfSale(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAvailableSlotsPointOfSale", reflect.TypeOf((*MockLowriBeckClient)(nil).GetAvailableSlotsPointOfSale), varargs...)
}
