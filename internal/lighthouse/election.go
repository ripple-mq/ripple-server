package lighthouse

import (
	"github.com/ripple-mq/ripple-server/internal/lighthouse/utils"
)

// StartElectLoop starts the election loop and watches for leader changes.
// It listens for leader signals and executes the provided callback when becoming the leader.
//
//	Async
func (t *LigthHouse) StartElectLoop(path utils.Path, data any, onBecommingLeaderCh chan<- struct{}) <-chan struct{} {
	t.elector.Start(path, data)
	go func() {
		for range t.elector.ListenForLeaderSignal() {
			onBecommingLeaderCh <- struct{}{}
		}
	}()
	return t.elector.ListenForFatalSignal()
}

// RegisterAsFollower registers `data` to the followers directory.
//
// It adds the provided data to the follower directory and returns the path of the registered follower.
// Returns the path and any error encountered during registration.
func (t *LigthHouse) RegisterAsFollower(path utils.Path, data any) (utils.Path, error) {
	path, err := t.elector.RegisterFollower(path, data)
	if err != nil {
		return utils.Path{}, err
	}
	return path, nil
}

// ReadLeader reads `data` from the leader directory.
//
// Retrieves the data associated with the leader at the specified path.
// Returns the leader data and any error encountered during reading
func (t *LigthHouse) ReadLeader(path utils.Path) ([]byte, error) {
	data, err := t.elector.ReadLeader(path)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// ReadFollowers reads data from followers.
//
// Retrieves the data associated with all followers at the specified path.
// Returns the followers' data and any error encountered during reading.
func (t *LigthHouse) ReadFollowers(path utils.Path) ([][]byte, error) {
	data, err := t.elector.ReadFollowers(path)
	if err != nil {
		return nil, err
	}
	return data, nil
}
