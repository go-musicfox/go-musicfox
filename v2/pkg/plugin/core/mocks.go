package plugin

import (
	"github.com/stretchr/testify/mock"
)

// MockEventBus Mock事件总线
type MockEventBus struct {
	mock.Mock
}

func (m *MockEventBus) Publish(event string, data interface{}) error {
	args := m.Called(event, data)
	return args.Error(0)
}

func (m *MockEventBus) Subscribe(event string, handler func(interface{})) error {
	args := m.Called(event, handler)
	return args.Error(0)
}

func (m *MockEventBus) Unsubscribe(event string, handler func(interface{})) error {
	args := m.Called(event, handler)
	return args.Error(0)
}

// MockServiceRegistry Mock服务注册表
type MockServiceRegistry struct {
	mock.Mock
}

func (m *MockServiceRegistry) RegisterService(name string, service interface{}) error {
	args := m.Called(name, service)
	return args.Error(0)
}

func (m *MockServiceRegistry) GetService(name string) (interface{}, error) {
	args := m.Called(name)
	return args.Get(0), args.Error(1)
}

func (m *MockServiceRegistry) UnregisterService(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

func (m *MockServiceRegistry) ListServices() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

// MockSecurityManager Mock安全管理器
type MockSecurityManager struct {
	mock.Mock
}

func (m *MockSecurityManager) ValidatePlugin(pluginPath string) error {
	args := m.Called(pluginPath)
	return args.Error(0)
}

func (m *MockSecurityManager) CheckPermissions(pluginID string, operation string) error {
	args := m.Called(pluginID, operation)
	return args.Error(0)
}

func (m *MockSecurityManager) GrantPermission(pluginID string, permission string) error {
	args := m.Called(pluginID, permission)
	return args.Error(0)
}

func (m *MockSecurityManager) RevokePermission(pluginID string, permission string) error {
	args := m.Called(pluginID, permission)
	return args.Error(0)
}