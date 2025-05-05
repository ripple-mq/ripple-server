package cmd_test

import (
	"os"
	"testing"
)

const (
	envProd  = "production_mode"
	envDebug = "debug_mode"
	envBuild = "__BUILD_MODE__"
)

func resetEnv() {
	// Unset all environment variables
	os.Unsetenv(envProd)
	os.Unsetenv(envDebug)
	os.Unsetenv(envBuild)
}

func TestMain(m *testing.M) {
	// Run all tests, unset env variables, and exit
	resetEnv()
	val := m.Run()
	resetEnv()
	os.Exit(val)
}
