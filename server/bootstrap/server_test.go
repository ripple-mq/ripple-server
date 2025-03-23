package server

import (
	"testing"
)

func TestNewBoostrapServer(t *testing.T) {
	type args struct {
		addr string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Test normal flow",
			args:    args{addr: ":9000"},
			wantErr: false,
		},
		{
			name:    "Test error flow",
			args:    args{addr: ":abc"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewBoostrapServer(tt.args.addr)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewBoostrapServer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
