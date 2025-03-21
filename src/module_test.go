package src

import (
	"testing"
)

func TestModuleName(t *testing.T) {
	if ProjectName() != "ripple-server" {
		t.Errorf("Project name `%s` incorrect", ProjectName())
	}
}
