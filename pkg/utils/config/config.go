package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	testutils "github.com/ripple-mq/ripple-server/test"
)

type Config struct {
	Zookeeper struct {
		Port                    int
		Connection_wait_time_ms int
		Session_timeout_ms      int
	}
	Server struct {
		Internal struct {
			Addr string
		}
		Exposed struct {
			Addr string
		}
	}
	Topic struct {
		Replicas int
	}
	Gossip struct {
		EventLoop struct {
			Max_fd_soft_limit      int64
			Task_queue_buffer_size int32
			Blocking_mode          bool
			Kqueue                 struct {
				Event_buffer_size int32
			}
			Epoll struct {
				Event_buffer_size int32
			}
		}
	}
	AsyncTCP struct {
		Addr      string
		EventLoop struct {
			Max_fd_soft_limit      int64
			Task_queue_buffer_size int32
			Blocking_mode          bool
			Kqueue                 struct {
				Event_buffer_size int32
			}
			Epoll struct {
				Event_buffer_size int32
			}
		}
	}
}

var Conf, _ = LoadConfig(".")

func LoadConfig(path string) (*Config, error) {
	testutils.SetRoot()
	var config Config
	if _, err := toml.DecodeFile("config.toml", &config); err != nil {
		fmt.Println("Error decoding TOML:", err)
		os.Exit(1)
	}
	return &config, nil
}
