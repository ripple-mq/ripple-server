package lighthouse

import (
	"github.com/ripple-mq/ripple-server/internal/lighthouse/utils"
)

func (t *LigthHouse) StartElectLoop(path utils.Path, data any, onBecommingLeader func(path utils.Path)) <-chan struct{} {
	t.elector.Start(path, data)
	go func() {
		for range t.elector.ListenForLeaderSignal() {
			go onBecommingLeader(path)
		}
	}()
	return t.elector.ListenForFatalSignal()
}

func (t *LigthHouse) RegisterAsFollower(path utils.Path, data any) (utils.Path, error) {
	path, err := t.elector.RegisterFollower(path, data)
	if err != nil {
		return utils.Path{}, err
	}
	return path, nil
}
