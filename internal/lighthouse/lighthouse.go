package lighthouse

import (
	"log"

	"github.com/ripple-mq/ripple-server/internal/lighthouse/election"
	"github.com/ripple-mq/ripple-server/internal/lighthouse/io"
	u "github.com/ripple-mq/ripple-server/internal/lighthouse/utils"
)

type LigthHouse struct {
	elector *election.LeaderElection
	io      *io.IO
}

var ligthHouseInstance *LigthHouse

func GetLightHouse() (*LigthHouse, error) {
	if ligthHouseInstance != nil {
		return ligthHouseInstance, nil
	}
	return new(), nil
}

func new() *LigthHouse {
	ioInstance := io.NewIO()
	elector := election.NewLeaderElection(ioInstance)
	ligthHouseInstance = &LigthHouse{io: io.NewIO(), elector: elector}
	return ligthHouseInstance
}

func (t *LigthHouse) EnsurePathExists(path string) {
	if err := t.io.EnsurePathExists(path); err != nil {
		log.Fatal(err)
	}
}

func (t *LigthHouse) RegisterSequential(path u.Path, data interface{}) u.Path {
	path, err := t.io.RegisterSequential(path, data)
	if err != nil {
		log.Fatal(err)
	}
	return path
}

func (t *LigthHouse) Read(path u.Path) ([]byte, error) {
	return t.io.Read(path)
}

func (t *LigthHouse) Write(path u.Path, newData any) {
	t.io.Write(path, newData)
}
