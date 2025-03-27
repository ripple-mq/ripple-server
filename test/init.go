package testutils

import (
	"os"
	"path/filepath"
)

func SetRoot() {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			panic("Project root not found")
		}
		dir = parent
	}
	os.Chdir(dir)
}
