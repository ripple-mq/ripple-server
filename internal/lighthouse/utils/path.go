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
	pwdI int
	cmp  []string
}

func (t PathBuilder) Base(p Path) PathBuilder {
	t.cmp = slices.Clone(p.Cmp)
	t.pwdI = len(p.Cmp)
	return t
}

func (t PathBuilder) CD(s string) PathBuilder {
	t.cmp = append(t.cmp, s)
	return t
}

func (t PathBuilder) CDBack() PathBuilder {
	if len(t.cmp) > 0 {
		t.cmp = t.cmp[:len(t.cmp)-1]
	}
	return t
}

func (t PathBuilder) GetDir() string {
	path := strings.Join(t.cmp, "/")
	return fmt.Sprintf("/%s/", path)
}

func (t PathBuilder) GetFile() string {
	path := strings.Join(t.cmp, "/")
	return fmt.Sprintf("/%s", path)
}

func (t PathBuilder) FileName() string {
	if len(t.cmp) != 0 {
		return t.cmp[len(t.cmp)-1]
	}
	return ""
}

func (t PathBuilder) Create() Path {
	return Path{Cmp: t.cmp}
}
