package config

import (
	"fmt"

	testutils "github.com/ripple-mq/ripple-server/test"
	"github.com/spf13/viper"
)

type Config struct {
	Zookeeper struct {
		Port                    int
		Connection_wait_time_ms int
		Session_timeout_ms      int
	}
	Server struct {
		Internal_grpc_addr string
		Exposed_grpc_addr  string
		Client_grpc_addr   string
	}
	Broker struct {
		ackLevel int
	}
	Topic struct {
		Replicas int
	}
	EventLoop struct {
		Max_fd_soft_limit          int64
		Task_queue_buffer_size     int32
		Kqueue_event_buffer_size   int32
		Epoll_event_buffer_size    int32
		Blocking_mode              bool
		Write_time_space_ms        int32
		Max_connection             int32
		Per_connection_buffer_size int32
	}
	AsyncTCP struct {
		Port int
	}
}

var Conf, _ = LoadConfig(".")

func LoadConfig(path string) (*Config, error) {
	testutils.SetRoot()
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(path)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %w", err)
	}

	return &config, nil
}
