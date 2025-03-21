package src_test

import (
	"testing"

	src "github.com/ripple.mq/ripple-server/internal"
)

func TestModuleName(t *testing.T) {
	if src.ProjectName() != "ripple-server" {
		t.Errorf("Project name `%s` incorrect", src.ProjectName())
	}
}
