package src_test

import (
	"testing"

	src "github.com/ripple-mq/ripple-server/internal"
	testutils "github.com/ripple-mq/ripple-server/test"
)

func TestModuleName(t *testing.T) {
	testutils.SetRoot()
	if src.ProjectName() != "ripple-server" {
		t.Errorf("Project name `%s` incorrect", src.ProjectName())
	}
}
