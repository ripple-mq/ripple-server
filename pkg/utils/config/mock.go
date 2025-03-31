package config

import "github.com/mohae/deepcopy"

type MockConfig struct {
	copy Config
}

var mockConfigInstance *MockConfig

func GetMockConfig() *MockConfig {
	if mockConfigInstance == nil {
		mockConfigInstance = new()
	}
	return mockConfigInstance
}

func new() *MockConfig {
	return &MockConfig{copy: deepcopy.Copy(*Conf).(Config)}
}

func (t *MockConfig) Reset() {
	Conf = &t.copy
}
