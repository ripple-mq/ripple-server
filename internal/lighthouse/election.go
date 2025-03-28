package lighthouse

import (
	"github.com/ripple-mq/ripple-server/internal/lighthouse/utils"
)

// StartElectLoop starts election loop and watching leader
//
//	Async
//
// Parameters:
//   - path(utils.Path): Base path containing followers & leader
//   - data(any): data to be written once becoming leader
//   - func(path utils.Path): function to execute once becoming leader (Async)
func (t *LigthHouse) StartElectLoop(path utils.Path, data any, onBecommingLeader func(path utils.Path)) <-chan struct{} {
	t.elector.Start(path, data)
	go func() {
		for range t.elector.ListenForLeaderSignal() {
			go onBecommingLeader(path)
		}
	}()
	return t.elector.ListenForFatalSignal()
}

// RegisterAsFollower registers `data` to followers directory
//
// Parameters:
//   - path(utils.Path): Base path
//   - data(any): data to be written
//
// Returns:
//   - uitls.Path
//   - error
func (t *LigthHouse) RegisterAsFollower(path utils.Path, data any) (utils.Path, error) {
	path, err := t.elector.RegisterFollower(path, data)
	if err != nil {
		return utils.Path{}, err
	}
	return path, nil
}

// ReadLeader reads `data` from leader directory
//
// Parameters:
//   - path(utils.Path): Base path
//
// Returns:
//   - []byte
//   - error
func (t *LigthHouse) ReadLeader(path utils.Path) ([]byte, error) {
	data, err := t.elector.ReadLeader(path)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// ReadFollowers reads data from followers
//
// Parameters:
//   - path(utils.Path): Base path
//
// Returns:
//   - [][]byte
//   - error
func (t *LigthHouse) ReadFollowers(path utils.Path) ([][]byte, error) {
	data, err := t.elector.ReadFollowers(path)
	if err != nil {
		return nil, err
	}
	return data, nil
}
