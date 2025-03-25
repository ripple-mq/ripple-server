package utils

import (
	"fmt"
	"slices"
	"strings"
)

type Path struct {
	Cmp []string
}

func Root() Path {
	return Path{Cmp: []string{}}
}

type PathBuilder struct {
	Cmp []string
}

func (t PathBuilder) Base(p Path) PathBuilder {
	t.Cmp = slices.Clone(p.Cmp)
	return t
}

func (t PathBuilder) CD(s string) PathBuilder {
	t.Cmp = append(t.Cmp, s)
	return t
}

func (t PathBuilder) CDBack() PathBuilder {
	if len(t.Cmp) > 0 {
		t.Cmp = t.Cmp[:len(t.Cmp)-1]
	}
	return t
}

func (t PathBuilder) GetDir() string {
	path := strings.Join(t.Cmp, "/")
	return fmt.Sprintf("/%s/", path)
}

func (t PathBuilder) GetFile() string {
	path := strings.Join(t.Cmp, "/")
	return fmt.Sprintf("/%s", path)
}

func (t PathBuilder) FileName() string {
	if len(t.Cmp) != 0 {
		return t.Cmp[len(t.Cmp)-1]
	}
	return ""
}

func (t PathBuilder) Create() Path {
	return Path(t)
}
