package server

import (
	"fmt"
	"math/rand"
	"testing"
)

const (
	minPort = 1024
	maxPort = 49150
)

func RandLocalAddr() string {
	randomNumber := rand.Intn(maxPort-minPort) + minPort
	return fmt.Sprintf(":%d", randomNumber)
}

func TesNewBootstrapServer(t *testing.T) {
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
			args:    args{addr: RandLocalAddr()},
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
			_, err := NewBootstrapServer(tt.args.addr)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewBoostrapServer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
