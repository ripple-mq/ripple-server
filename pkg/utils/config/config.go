package config

import (
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	testutils "github.com/ripple-mq/ripple-server/test"
	"github.com/spf13/viper"
)

type Config struct {
	Zookeeper struct {
		Connection_url          string
		Connection_wait_time_ms int
		Session_timeout_ms      int
	}
	Server struct {
		Internal_grpc_port int
		Exposed_grpc_port  int
	}
}

var Conf, _ = LoadConfig(".")

func LoadConfig(path string) (*Config, error) {
	testutils.SetRoot()
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(path)
	viper.AutomaticEnv()

	pwd, _ := os.Getwd()
	log.Infof("Present INNNN:  %s", pwd)

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %w", err)
	}

	return &config, nil
}
