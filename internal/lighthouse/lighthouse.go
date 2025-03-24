package lighthouse

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/ripple-mq/ripple-server/internal/lighthouse/utils"
	"github.com/samuel/go-zookeeper/zk"
)

type LigthHouse struct {
	conn *zk.Conn
}

var ligthHouseInstance *LigthHouse

func GetLightHouse() (*LigthHouse, error) {
	if ligthHouseInstance != nil {
		return ligthHouseInstance, nil
	}
	ligthHouseInstance = &LigthHouse{}
	ligthHouseInstance.Connect()
	return ligthHouseInstance, nil
}

func (t *LigthHouse) Connect() {
	if t.conn != nil && t.conn.State() == zk.StateHasSession {
		return
	}
	conn, _, err := zk.Connect([]string{"localhost:2181"}, time.Second)
	if err != nil {
		log.Fatal("Failed to connect to Zookeeper:", err)
	}
	time.Sleep(1 * time.Second)
	t.conn = conn
}

func (t *LigthHouse) EnsurePathExists(path string) error {
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

func (t *LigthHouse) RegisterSequential(path Path, data interface{}) Path {
	err := t.EnsurePathExists(path.BasePath())
	if err != nil {
		log.Errorf("Failed to register: %v", err)
		return Path{}
	}
	ephemeralNodePath, err := t.conn.CreateProtectedEphemeralSequential(path.BasePath()+"/", utils.ToBytes(data), zk.WorldACL(zk.PermAll))
	if err != nil {
		return Path{}
	}
	pathSplits := strings.Split(ephemeralNodePath, "/")
	return Path{Base: path.Base, Role: path.Role, Name: pathSplits[len(pathSplits)-1]}
}

func (t *LigthHouse) Read(path Path) ([]byte, error) {
	err := t.EnsurePathExists(path.BasePath())
	if err != nil {
		return nil, fmt.Errorf("failed to read: %v", err)
	}
	data, _, err := t.conn.Get(path.FullPath())
	if err != nil {
		return nil, fmt.Errorf("error reading znode: %v", err)
	}
	return data, nil
}

func (t *LigthHouse) UpdateZnode(path Path, newData any) {
	_, stat, err := t.conn.Get(path.FullPath())
	if err != nil {
		log.Fatalf("Failed to read znode before updating: %v", err)
	}

	_, err = t.conn.Set(path.FullPath(), utils.ToBytes(newData), stat.Version)
	if err != nil {
		log.Fatalf("Failed to update znode: %v", err)
	}
}
