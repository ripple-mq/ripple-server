package io

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	u "github.com/ripple-mq/ripple-server/internal/lighthouse/utils"
	"github.com/ripple-mq/ripple-server/pkg/utils/config"
	"github.com/ripple-mq/ripple-server/pkg/utils/env"
	"github.com/samuel/go-zookeeper/zk"
)

const localAddr string = "127.0.0.1"

type IO struct {
	conn *zk.Conn
}

var ioInstance *IO

func GetIO() *IO {
	if ioInstance != nil {
		return ioInstance
	}
	return newIO()
}

func newIO() *IO {
	return &IO{connect()}
}

// connect establishes a connection to Zookeeper using the configuration settings.
//
// It returns a pointer to a Zookeeper connection. If the connection fails, the application
// will log a fatal error and terminate.
func connect() *zk.Conn {
	url := fmt.Sprintf("%s:%d", env.Get("ZK_IPv4", localAddr), config.Conf.Zookeeper.Port)
	log.Infof("Attempting zookeeper connection: %s", url)
	conn, _, err := zk.Connect([]string{url}, time.Duration(config.Conf.Zookeeper.Session_timeout_ms*int(time.Millisecond)))
	if err != nil {
		log.Fatal("Failed to connect to Zookeeper:", err)
	}
	time.Sleep(time.Duration(config.Conf.Zookeeper.Connection_wait_time_ms * int(time.Millisecond)))
	return conn
}

// EnsurePathExists checks if the specified `path` exists in Zookeeper,
// and creates it if it does not.
//
// It traverses the path segments and creates each intermediate directory if needed.
func (t *IO) EnsurePathExists(path string) error {
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return fmt.Errorf("invalid path: %s", path)
	}

	fullPath := ""
	for _, part := range parts {
		if part == "" {
			continue
		}
		fullPath += "/" + part

		exists, _, err := t.conn.Exists(fullPath)
		if err != nil {
			return fmt.Errorf("error checking path %s: %v", fullPath, err)
		}

		if !exists {
			_, err := t.conn.Create(fullPath, []byte{}, 0, zk.WorldACL(zk.PermAll))
			if err != nil && err != zk.ErrNodeExists {
				return fmt.Errorf("error creating path %s: %v", fullPath, err)
			}
		}
	}
	return nil
}

// RegisterSequential registers a sequential ephemeral node at the specified path.
//
// Note: The node will be automatically removed when the connection goes down.
//
// It ensures the existence of the parent path and creates a sequential ephemeral node
// with the provided data.
func (t *IO) RegisterSequential(path u.Path, data any) (u.Path, error) {
	err := t.EnsurePathExists(u.PathBuilder{}.Base(path).GetDir())
	if err != nil {
		return u.Path{}, fmt.Errorf("failed to register: %v", err)
	}
	ephemeralNodePath, err := t.conn.CreateProtectedEphemeralSequential(u.PathBuilder{}.Base(path).GetDir(), u.ToBytes(data), zk.WorldACL(zk.PermAll))
	if err != nil {
		return u.Path{}, nil
	}

	return u.Path{Cmp: strings.Split(ephemeralNodePath, "/")[1:]}, nil
}

// Read retrieves data from the specified path.
//
// Note: If the path doesn't exist, it will be created.
//
// This function ensures the existence of the path and then reads the data from the
// given node in the Lighthouse.
func (t *IO) Read(path u.Path) ([]byte, error) {
	err := t.EnsurePathExists(u.PathBuilder{}.Base(path).GetDir())
	if err != nil {
		return nil, fmt.Errorf("failed to read: %v", err)
	}
	data, _, err := t.conn.Get(u.PathBuilder{}.Base(path).GetFile())
	if err != nil {
		return nil, fmt.Errorf("error reading node: %v", err)
	}
	return data, nil
}

// Write stores data at the specified path.
//
// Note: If the path does not exist, it will be created before writing the data.
//
// This function first checks if the path exists and then updates the node with
// the new data. If the node doesn’t exist, it creates the path and writes the data.
func (t *IO) Write(path u.Path, newData any) error {
	_, stat, err := t.conn.Get(u.PathBuilder{}.GetFile())
	if err != nil {
		return fmt.Errorf("failed to read before updating: %v", err)
	}

	_, err = t.conn.Set(u.PathBuilder{}.Base(path).GetFile(), u.ToBytes(newData), stat.Version)
	if err != nil {
		return fmt.Errorf("failed to update znode: %v", err)
	}
	return nil
}

// GetChildrenAndWatch retrieves the list of children at the given path and
// returns a channel to watch for changes.
func (t *IO) GetChildrenAndWatch(path u.Path) ([]string, <-chan zk.Event, error) {
	children, _, ch, err := t.conn.ChildrenW(u.PathBuilder{}.Base(path).GetFile())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get & watch : %v %s", err, path)
	}
	return children, ch, nil
}

// GetChildren retrieves the list of children at the specified path.
func (t *IO) GetChildren(path u.Path) ([]string, error) {
	children, _, err := t.conn.Children(u.PathBuilder{}.Base(path).GetFile())
	if err != nil || len(children) == 0 {
		return nil, fmt.Errorf("failed to get childrens: %v %s", err, path)
	}
	return children, nil
}
