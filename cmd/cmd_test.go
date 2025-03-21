package cmd_test

import (
	"os"
	"testing"

	"github.com/ripple-mq/ripple-server/cmd"
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

func TestMainMethod(_ *testing.T) {
	cmd.Execute()
}

func dirtyCheck(t *testing.T, val01, val02 bool) {
	t.Helper()

	if val01 != val02 {
		t.Errorf("Values don't match: (%v, %v)", val01, val02)
	}
}

func TestFirst(t *testing.T) {
	resetEnv()
	dirtyCheck(t, cmd.FirstCheck(), false) // no variable should be detected

	os.Setenv(envProd, "production")
	dirtyCheck(t, cmd.FirstCheck(), true) // detect `production` mode

	resetEnv()
	os.Setenv(envDebug, "debug")
	dirtyCheck(t, cmd.FirstCheck(), true) // detect `debug` mode
}

func TestSecond(t *testing.T) {
	resetEnv()
	dirtyCheck(t, cmd.SecondCheck(), false) // no variable should be detected

	resetEnv()
	os.Setenv(envBuild, "production")
	dirtyCheck(t, cmd.SecondCheck(), true) // detect `production` mode!

	resetEnv()
	os.Setenv(envBuild, "debug")
	dirtyCheck(t, cmd.SecondCheck(), true) // detect `debug` mode!
}
