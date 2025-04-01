package asynctcp_test

import (
	"fmt"
	"math/rand"
	"net"
	"testing"
	"time"

	"github.com/charmbracelet/log"
	"github.com/ripple-mq/ripple-server/internal/lighthouse"
	"github.com/ripple-mq/ripple-server/pkg/p2p/encoder"
	"github.com/ripple-mq/ripple-server/pkg/p2p/transport/asynctcp"
	"github.com/ripple-mq/ripple-server/pkg/p2p/transport/asynctcp/comm"
	"github.com/ripple-mq/ripple-server/pkg/p2p/transport/tcp"
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
		id      string
		wantErr bool
	}{
		{
			name:    "Invalid Address - Missing Port",
			id:      "127.0.0.1",
			wantErr: false,
		},
		{
			name:    "Valid Address",
			id:      "hi",
			wantErr: false,
		},
		{
			name:    "Invalid Address - Nonsense String",
			id:      "some id",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := asynctcp.NewTransport(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTransport() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestTransport_Listen(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "Valid Transport Listening",
			id:      "some id",
			wantErr: false,
		},
		{
			name:    "Invalid Transport - Bad Address",
			id:      "some id",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr, err := asynctcp.NewTransport(tt.id)
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

func TestTransport_Recieve(t *testing.T) {
	tests := []struct {
		name       string
		serverId   string
		clientAddr string
		data       string
		wantErr    bool
	}{
		{
			name:       "Valid Transport Listening",
			serverId:   "123",
			clientAddr: RandLocalAddr(),
			data:       "sample data",
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lighthouse.GetLightHouse()
			server, _ := asynctcp.NewTransport(tt.serverId, asynctcp.TransportOpts{OnAcceptingConn: func(msg comm.Message) { fmt.Println(msg) }})
			client, err := tcp.NewTransport(tt.clientAddr, dummyOnConnFunction)

			if err != nil {
				if !tt.wantErr {
					t.Errorf("NewTransport() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			err = server.Listen()
			if (err != nil) != tt.wantErr {
				t.Errorf("Transport.Listen() error = %v, wantErr %v", err, tt.wantErr)
			}
			go func() {
				for {
					var msg string
					server.Consume(encoder.GOBDecoder{}, &msg)
					fmt.Println("Receieved:  ", msg)
					if msg != tt.data {
						t.Errorf("Transport.Listen() got = %s, want = %s ", msg, tt.data)
					}
				}
			}()

			time.Sleep(1000 * time.Millisecond)

			if err := client.SendToAsync(tt.serverId, "some metadat", tt.data); (err != nil) != tt.wantErr {
				t.Errorf("Transport.Send() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := client.SendToAsync(tt.serverId, struct{}{}, tt.data); (err != nil) != tt.wantErr {
				t.Errorf("Transport.Send() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := client.SendToAsync(tt.serverId, struct{}{}, tt.data); (err != nil) != tt.wantErr {
				t.Errorf("Transport.Send() error = %v, wantErr %v", err, tt.wantErr)
			}
			time.Sleep(1000 * time.Millisecond)
		})
	}
}

func TestTransport_Send(t *testing.T) {
	tests := []struct {
		name       string
		serverAddr string
		clientId   string
		data       string
		wantErr    bool
	}{
		{
			name:       "Valid Transport Listening",
			serverAddr: RandLocalAddr(),
			clientId:   "1526",
			data:       "sample data",
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, _ := tcp.NewTransport(tt.serverAddr, dummyOnConnFunction)
			client, err := asynctcp.NewTransport(tt.clientId)

			if err != nil {
				if !tt.wantErr {
					t.Errorf("NewTransport() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			err = server.Listen()
			if (err != nil) != tt.wantErr {
				t.Errorf("Transport.Listen() error = %v, wantErr %v", err, tt.wantErr)
			}
			go func() {
				var msg string
				_, _ = server.Consume(encoder.GOBDecoder{}, &msg)
				fmt.Println("Receieved:  ", msg)
				if msg != tt.data {
					t.Errorf("Transport.Listen() got = %s, want = %s ", msg, tt.data)
				}
			}()

			time.Sleep(1000 * time.Millisecond)

			if err := client.Send(tt.serverAddr, "some metadata", tt.data); err != nil {
				fmt.Println("error() , ", err)
			}

			time.Sleep(1000 * time.Millisecond)
		})
	}
}
