package utils_test

import (
	"reflect"
	"testing"

	"github.com/ripple-mq/ripple-server/internal/lighthouse/utils"
	testutils "github.com/ripple-mq/ripple-server/test"
)

func TestPathBuilder_Base(t *testing.T) {
	testutils.SetRoot()
	type fields struct {
		Cmp []string
	}
	type args struct {
		p utils.Path
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   utils.PathBuilder
	}{
		{
			name:   "Valid non root as base",
			fields: fields{},
			args:   args{utils.Path{Cmp: []string{"base"}}},
			want:   utils.PathBuilder{Cmp: []string{"base"}},
		},
		{
			name:   "Valid root as base",
			fields: fields{},
			args:   args{utils.Root()},
			want:   utils.PathBuilder{Cmp: []string{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := utils.PathBuilder{}
			if got := tr.Base(tt.args.p); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PathBuilder.Base() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPathBuilder_CD(t *testing.T) {
	testutils.SetRoot()
	type fields struct {
		Cmp []string
	}
	type args struct {
		s string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   utils.PathBuilder
	}{
		{
			name:   "valid CD new dir",
			fields: fields{Cmp: []string{"base"}},
			args:   args{s: "newDir"},
			want:   utils.PathBuilder{Cmp: []string{"base", "newDir"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := utils.PathBuilder{
				Cmp: tt.fields.Cmp,
			}
			if got := tr.CD(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PathBuilder.CD() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPathBuilder_CDBack(t *testing.T) {
	testutils.SetRoot()
	type fields struct {
		Cmp []string
	}
	tests := []struct {
		name   string
		fields fields
		want   utils.PathBuilder
	}{
		{
			name:   "cd newDir",
			fields: fields{Cmp: []string{"base", "newDir"}},
			want:   utils.PathBuilder{Cmp: []string{"base"}},
		},
		{
			name:   "cd .. at root",
			fields: fields{Cmp: []string{}},
			want:   utils.PathBuilder{Cmp: []string{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := utils.PathBuilder{
				Cmp: tt.fields.Cmp,
			}
			if got := tr.CDBack(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PathBuilder.CDBack() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPathBuilder_FileName(t *testing.T) {
	testutils.SetRoot()
	type fields struct {
		Cmp []string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "Valid FileName",
			fields: fields{Cmp: []string{"base", "newDir"}},
			want:   "newDir",
		},
		{
			name:   "File name at root",
			fields: fields{Cmp: []string{}},
			want:   "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := utils.PathBuilder{
				Cmp: tt.fields.Cmp,
			}
			if got := tr.FileName(); got != tt.want {
				t.Errorf("PathBuilder.FileName() = %v, want %v", got, tt.want)
			}
		})
	}
}
