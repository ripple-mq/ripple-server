package lighthouse

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/ripple-mq/ripple-server/internal/lighthouse/utils"
	"github.com/ripple-mq/ripple-server/pkg/utils/config"
	"github.com/samuel/go-zookeeper/zk"
)

type LigthHouse struct {
	conn *zk.Conn
}

var ligthHouseInstance, _ = newLightHouse()

func GetLightHouse() (*LigthHouse, error) {
	return ligthHouseInstance, nil
}

func newLightHouse() (*LigthHouse, error) {

	conn, _, err := zk.Connect([]string{config.Conf.Zookeeper.Address}, time.Second)
	if err != nil {
		log.Fatal("Failed to connect to Zookeeper:", err)
	}

	time.Sleep(1 * time.Second)

	state := conn.State()
	if state != zk.StateHasSession {
		log.Warnf("Zookeeper connection state: %v", state)
	} else {
		log.Info("Successfully connected to Zookeeper")
	}

	return &LigthHouse{conn}, nil
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
