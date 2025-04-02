package eventloop_test

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"syscall"
	"testing"
	"time"

	"github.com/ripple-mq/ripple-server/internal/gossip/eventloop"
	"github.com/ripple-mq/ripple-server/pkg/p2p/encoder"
	"github.com/ripple-mq/ripple-server/pkg/p2p/transport/tcp"
	"github.com/ripple-mq/ripple-server/pkg/utils"
	"github.com/ripple-mq/ripple-server/pkg/utils/collection"
	"github.com/ripple-mq/ripple-server/pkg/utils/config"
)

func TestGetEventLoop(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "normal flow",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := eventloop.GetEventLoop()
			if (err != nil) != tt.wantErr || got == nil {
				t.Errorf("GetEventLoop() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

type Task struct {
	AckCount *collection.ConcurrentValue[int] // counts ack for testing purpose
	conn     net.Conn
}

func (t Task) GetFD() int {
	syscallConn, ok := t.conn.(syscall.Conn)
	if !ok {
		return 0
	}

	rawConn, err := syscallConn.SyscallConn()
	if err != nil {
		return 0
	}

	var fd uintptr
	_ = rawConn.Control(func(fdParam uintptr) {
		fd = fdParam
	})

	return int(fd)
}

func (t Task) GetDataToSend() []byte {
	var msg bytes.Buffer
	_ = encoder.GOBEncoder{}.Encode("hi from client", &msg)

	length := uint32(len(msg.Bytes()))
	lengthBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBytes, length)

	lengthBytes = append(lengthBytes, msg.Bytes()...)
	return lengthBytes
}

func (t Task) Acknowledge(data []byte) error {
	var msg string
	err := encoder.GOBDecoder{}.Decode(bytes.NewBuffer(data), &msg)
	if err != nil {
		fmt.Println("decode() err: ", err)
	}
	fmt.Println("Acknowledge: ", msg)
	t.AckCount.Set(t.AckCount.Get() + 1)
	return nil
}

// TestEventLoopIntegration verifies the integration of the event loop with
// TCP-based client-server communication.
// It tests TCP pings with multiple servers and counts the expected number of acknowledgments (acks).
func TestEventLoopIntegration(t *testing.T) {

	tests := []struct {
		name           string
		servers        []string
		clientAddr     string
		taskBufferSize int32
		tasksPerClient int
		wantGetLoopErr bool
		wantAddTaskErr bool
	}{
		{
			name: "normal flow",
			servers: []string{
				"127.0.0.1" + utils.RandLocalAddr(),
				"127.0.0.1" + utils.RandLocalAddr(),
				"127.0.0.1" + utils.RandLocalAddr(),
			},
			clientAddr:     ":8800",
			tasksPerClient: 30,
			taskBufferSize: 150,
			wantGetLoopErr: false,
			wantAddTaskErr: false,
		},
		{
			name: "buffer  overflow",
			servers: []string{
				"127.0.0.1" + utils.RandLocalAddr(),
			},
			clientAddr:     ":8801",
			tasksPerClient: 30,
			taskBufferSize: 10,
			wantGetLoopErr: false,
			wantAddTaskErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConf := config.GetMockConfig()
			defer mockConf.Reset()

			config.Conf.Gossip.EventLoop.Task_queue_buffer_size = tt.taskBufferSize

			for _, server := range tt.servers {
				server, _ := tcp.NewTransport(server, func(conn net.Conn, msg []byte) {})
				if err := server.Listen(); err != nil {
					t.Errorf("Listen() error= %v", err)
				}

				go func() {
					for {
						var msg string
						addr, _ := server.Consume(encoder.GOBDecoder{}, &msg)
						server.Send(addr, struct{}{}, "ACK=Done")
					}
				}()
			}

			// ShouldClientHandleConn is set to false to avoid multiple reads over same connection
			client, _ := tcp.NewTransport(tt.clientAddr, func(conn net.Conn, msg []byte) {}, tcp.TransportOpts{ShouldClientHandleConn: false})

			for _, server := range tt.servers {
				// TCP handshake
				if err := client.Send(server, struct{}{}, "some random data"); err != nil {
					fmt.Println("Send() err: ", err)
				}
			}

			el, err := eventloop.GetEventLoop()

			if (err != nil) != tt.wantGetLoopErr {
				t.Errorf("GetEventLoop() error = %v, wantErr %v", err, tt.wantGetLoopErr)
				return
			}

			el.StartExecLoop()
			el.StartAckLoop()

			for _, server := range tt.servers {
				ackCount := collection.NewConcurrentValue(0)
				task := Task{conn: client.PeersMap[server].GetConnection(), AckCount: ackCount}
				for range tt.tasksPerClient {
					if err := el.AddTask(task); err != nil && !tt.wantAddTaskErr {
						t.Errorf("failed to add task, %v", err)
					}
				}

				time.Sleep(2 * time.Second)
				if !tt.wantAddTaskErr && ackCount.Get() != tt.tasksPerClient+1 {
					t.Errorf("invalid ack count, got= %d , want= %d", ackCount.Get(), tt.tasksPerClient)
				}

			}

		})
	}
}
