package io

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	u "github.com/ripple-mq/ripple-server/internal/lighthouse/utils"
	"github.com/ripple-mq/ripple-server/pkg/utils/config"
	"github.com/samuel/go-zookeeper/zk"
)

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

// connect connects to zookeeper
func connect() *zk.Conn {
	conn, _, err := zk.Connect([]string{config.Conf.Zookeeper.Connection_url}, time.Duration(config.Conf.Zookeeper.Session_timeout_ms*int(time.Millisecond)))
	if err != nil {
		log.Fatal("Failed to connect to Zookeeper:", err)
	}
	time.Sleep(time.Duration(config.Conf.Zookeeper.Connection_wait_time_ms * int(time.Millisecond)))
	return conn
}

// EnsurePathExists checks whether `path` exists or not, if not create one
//
// Parameters:
//   - path(string)
//
// Returns:
//   - error
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
	log.Infof("PATH: %s", fullPath)
	return nil
}

// RegisterSequential registers sequential ephemeral node
//
// Note: It will be removed automatically when connection goes down
//
// Parameters:
//   - path(u.Path)
//   - data(any)
//
// Returns:
//   - u.Path
//   - error
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

// Read data from given path
//
// Note: it will create path if doesn't exists
func (t *IO) Read(path u.Path) ([]byte, error) {
	err := t.EnsurePathExists(u.PathBuilder{}.Base(path).GetDir())
	if err != nil {
		return nil, fmt.Errorf("failed to read: %v", err)
	}
	data, _, err := t.conn.Get(u.PathBuilder{}.Base(path).GetFile())
	if err != nil {
		return nil, fmt.Errorf("error reading znode: %v", err)
	}
	return data, nil
}

// Write data at given path
//
// Note: it will create path if doesn't exists
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

// GetChildrenAndWatch returns list of childrens at given `path` & `ch` to watch changes
func (t *IO) GetChildrenAndWatch(path u.Path) ([]string, <-chan zk.Event, error) {
	children, _, ch, err := t.conn.ChildrenW(u.PathBuilder{}.Base(path).GetFile())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get & watch : %v %s", err, path)
	}
	return children, ch, nil
}

// GetChildrenAndWatch returns list of childrens at given `path`
func (t *IO) GetChildren(path u.Path) ([]string, error) {
	children, _, err := t.conn.Children(u.PathBuilder{}.Base(path).GetFile())
	if err != nil || len(children) == 0 {
		return nil, fmt.Errorf("failed to get childrens: %v %s", err, path)
	}
	return children, nil
}
