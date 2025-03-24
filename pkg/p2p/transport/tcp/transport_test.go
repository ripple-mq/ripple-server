package tcp

import (
	"fmt"
	"math/rand"
	"net"
	"testing"
	"time"

	"github.com/charmbracelet/log"
	"github.com/ripple-mq/ripple-server/pkg/p2p/encoder"
)

const (
	minPort = 1024
	maxPort = 49150
)

func RandLocalAddr() string {
	for {
		port := rand.Intn(maxPort-minPort) + minPort
		addr := fmt.Sprintf(":%d", port)
		ln, err := net.Listen("tcp", addr)
		if err == nil {
			ln.Close()
			return addr
		}
	}
}

func dummyOnConnFunction(conn net.Conn, msg []byte) {
	log.Infof("handshake with %v %s", conn, string(msg))
}

func TestNewTransport(t *testing.T) {
	tests := []struct {
		name    string
		addr    string
		wantErr bool
	}{
		{
			name:    "Valid Address",
			addr:    RandLocalAddr(),
			wantErr: false,
		},
		{
			name:    "Invalid Address - Missing Port",
			addr:    "127.0.0.1",
			wantErr: true,
		},
		{
			name:    "Invalid Address - Nonsense String",
			addr:    "invalid_address",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewTransport(tt.addr, dummyOnConnFunction)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTransport() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.ListenAddr.String() != tt.addr {
				t.Errorf("NewTransport() ListenAddr = %v, want %v", got.ListenAddr, tt.addr)
			}
			if !tt.wantErr && got.IncommingMsgQueue == nil {
				t.Errorf("NewTransport() IncommingMsgQueue should not be nil")
			}
			if !tt.wantErr && got.PeersMap == nil {
				t.Errorf("NewTransport() PeersMap should not be nil")
			}
		})
	}
}

func TestTransport_Listen(t *testing.T) {
	tests := []struct {
		name    string
		addr    string
		wantErr bool
	}{
		{
			name:    "Valid Transport Listening",
			addr:    RandLocalAddr(),
			wantErr: false,
		},
		{
			name:    "Invalid Transport - Bad Address",
			addr:    "invalid_address",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr, err := NewTransport(tt.addr, dummyOnConnFunction)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("NewTransport() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			err = tr.Listen()
			if (err != nil) != tt.wantErr {
				t.Errorf("Transport.Listen() error = %v, wantErr %v", err, tt.wantErr)
			}

			time.Sleep(100 * time.Millisecond)
			if err := tr.Stop(); (err != nil) != tt.wantErr {
				t.Errorf("Transport.Stop() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTransport_Send(t *testing.T) {
	tests := []struct {
		name       string
		serverAddr string
		clientAddr string
		data       string
		wantErr    bool
	}{
		{
			name:       "Valid Transport Listening",
			serverAddr: RandLocalAddr(),
			clientAddr: RandLocalAddr(),
			data:       "sample data",
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, _ := NewTransport(tt.serverAddr, dummyOnConnFunction)
			client, err := NewTransport(tt.clientAddr, dummyOnConnFunction)

			if err != nil {
				if !tt.wantErr {
					t.Errorf("NewTransport() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			go func() {
				err := server.Listen()
				if (err != nil) != tt.wantErr {
					t.Errorf("Transport.Listen() error = %v, wantErr %v", err, tt.wantErr)
				}
			}()

			if err := client.Send(tt.serverAddr, "dummy metadata", tt.data); (err != nil) != tt.wantErr {
				t.Errorf("Transport.Send() error = %v, wantErr %v", err, tt.wantErr)
			}
			time.Sleep(1000 * time.Millisecond)
		})
	}
}

func TestTransport_Close(t *testing.T) {
	tests := []struct {
		name       string
		serverAddr string
		clientAddr string
		data       string
		wantErr    bool
	}{
		{
			name:       "Valid Transport Listening",
			serverAddr: RandLocalAddr(),
			clientAddr: RandLocalAddr(),
			data:       "sample data",
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, _ := NewTransport(tt.serverAddr, dummyOnConnFunction)
			client, err := NewTransport(tt.clientAddr, dummyOnConnFunction)

			if err != nil {
				if !tt.wantErr {
					t.Errorf("NewTransport() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			go func() {
				err := server.Listen()
				if (err != nil) != tt.wantErr {
					t.Errorf("Transport.Listen() error = %v, wantErr %v", err, tt.wantErr)
				}
			}()

			time.Sleep(100 * time.Millisecond)

			if err := client.Send(tt.serverAddr, "dummy metadata", tt.data); (err != nil) != tt.wantErr {
				t.Errorf("Transport.Send() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err := client.Close(tt.serverAddr); (err != nil) != tt.wantErr {
				t.Errorf("Transport.Close() error = %v, wantErr %v", err, tt.wantErr)
			}

		})
	}
}

func TestTransport_Consume(t *testing.T) {
	tests := []struct {
		name       string
		serverAddr string
		clientAddr string
		data       string
		wantErr    bool
	}{
		{
			name:       "Valid Transport Listening",
			serverAddr: RandLocalAddr(),
			clientAddr: RandLocalAddr(),
			data:       "sample data",
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, _ := NewTransport(tt.serverAddr, dummyOnConnFunction)
			client, err := NewTransport(tt.clientAddr, dummyOnConnFunction)

			if err != nil {
				if !tt.wantErr {
					t.Errorf("NewTransport() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			go func() {
				err := server.Listen()
				if (err != nil) != tt.wantErr {
					t.Errorf("Transport.Listen() error = %v, wantErr %v", err, tt.wantErr)
				}
			}()

			time.Sleep(100 * time.Millisecond)

			if err := client.Send(tt.serverAddr, "dummy metadata", tt.data); (err != nil) != tt.wantErr {
				t.Errorf("Transport.Send() error = %v, wantErr %v", err, tt.wantErr)
			}

			go func() {
				for {
					var msg string
					_, err := server.Consume(encoder.GOBDecoder{}, &msg)
					if err != nil || tt.data != msg {
						t.Errorf("Transport.Consume() error = %v, wantErr %v", err, tt.data)
					}
				}
			}()

			time.Sleep(time.Second * 2)

		})
	}
}
