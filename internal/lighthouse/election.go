package lighthouse

import (
	"log"

	"github.com/ripple-mq/ripple-server/internal/lighthouse/utils"
)

func (t *LigthHouse) StartElectLoop(path utils.Path, data any, onBecommingLeader func(path utils.Path)) {
	t.elector.Start(path, data)
	go func() {
		for range t.elector.Signal() {
			go onBecommingLeader(path)
		}
	}()
}

func (t *LigthHouse) RegisterAsFollower(path utils.Path, data any) utils.Path {
	path, err := t.elector.RegisterFollower(path, data)
	if err != nil {
		log.Fatal(err)
	}
	return path
}
