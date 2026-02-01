package services

import (
	"context"
	"reflect"

	"go.uber.org/mock/gomock"
)

// InvokerAPIClient defines the interface for calling Invoker Control Plane APIs.
// This allows mocking in tests without using the actual Invoker server.
type InvokerAPIClient interface {
	// GetAuthURL requests an OAuth authorization URL from Invoker.
	GetAuthURL(ctx context.Context, codeChallenge, state, redirectURI string) (string, error)

	// ExchangeCode exchanges OAuth code for machine credentials.
	ExchangeCode(ctx context.Context, code, state, codeVerifier string) (*OAuthExchangeResponse, error)

	// RegisterMachine registers the machine with Invoker.
	RegisterMachine(ctx context.Context, machineJWT string, req MachineRegisterRequest) (*MachineRegisterResponse, error)
}

// MockInvokerAPIClient is a mock implementation of InvokerAPIClient.
type MockInvokerAPIClient struct {
	ctrl     *gomock.Controller
	recorder *MockInvokerAPIClientMockRecorder
}

type MockInvokerAPIClientMockRecorder struct {
	mock *MockInvokerAPIClient
}

func NewMockInvokerAPIClient(ctrl *gomock.Controller) *MockInvokerAPIClient {
	mock := &MockInvokerAPIClient{ctrl: ctrl}
	mock.recorder = &MockInvokerAPIClientMockRecorder{mock}
	return mock
}

func (m *MockInvokerAPIClient) EXPECT() *MockInvokerAPIClientMockRecorder {
	return m.recorder
}

func (m *MockInvokerAPIClient) GetAuthURL(ctx context.Context, codeChallenge, state, redirectURI string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAuthURL", ctx, codeChallenge, state, redirectURI)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockInvokerAPIClientMockRecorder) GetAuthURL(ctx, codeChallenge, state, redirectURI interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAuthURL", reflect.TypeOf((*MockInvokerAPIClient)(nil).GetAuthURL), ctx, codeChallenge, state, redirectURI)
}

func (m *MockInvokerAPIClient) ExchangeCode(ctx context.Context, code, state, codeVerifier string) (*OAuthExchangeResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ExchangeCode", ctx, code, state, codeVerifier)
	ret0, _ := ret[0].(*OAuthExchangeResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockInvokerAPIClientMockRecorder) ExchangeCode(ctx, code, state, codeVerifier interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ExchangeCode", reflect.TypeOf((*MockInvokerAPIClient)(nil).ExchangeCode), ctx, code, state, codeVerifier)
}

func (m *MockInvokerAPIClient) RegisterMachine(ctx context.Context, machineJWT string, req MachineRegisterRequest) (*MachineRegisterResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RegisterMachine", ctx, machineJWT, req)
	ret0, _ := ret[0].(*MachineRegisterResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockInvokerAPIClientMockRecorder) RegisterMachine(ctx, machineJWT, req interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RegisterMachine", reflect.TypeOf((*MockInvokerAPIClient)(nil).RegisterMachine), ctx, machineJWT, req)
}

// testableOAuthService extends OAuthService with injectable Invoker client for testing.
type testableOAuthService struct {
	*OAuthService
	invokerClient InvokerAPIClient
}

func (s *testableOAuthService) callInvokerAuthURL(ctx context.Context, codeChallenge, state, redirectURI string) (string, error) {
	return s.invokerClient.GetAuthURL(ctx, codeChallenge, state, redirectURI)
}

func (s *testableOAuthService) callInvokerExchange(ctx context.Context, code, state, codeVerifier string) (*OAuthExchangeResponse, error) {
	return s.invokerClient.ExchangeCode(ctx, code, state, codeVerifier)
}

func (s *testableOAuthService) callInvokerRegisterMachine(ctx context.Context, machineJWT string, req MachineRegisterRequest) (*MachineRegisterResponse, error) {
	return s.invokerClient.RegisterMachine(ctx, machineJWT, req)
}
