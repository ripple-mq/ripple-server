package lighthouse

import (
	"fmt"
	"log"
	"sync"

	"github.com/ripple-mq/ripple-server/internal/lighthouse/election"
	"github.com/ripple-mq/ripple-server/internal/lighthouse/io"
	u "github.com/ripple-mq/ripple-server/internal/lighthouse/utils"
)

func init() {
	newLH()
}

// LigthHouse holds io.IO & election.LeaderElectio instance
type LigthHouse struct {
	elector *election.LeaderElection
	io      *io.IO
}

var (
	ligthHouseInstance *LigthHouse
	once               sync.Once
)

// GetLightHouse returns singleton instance of *LigthHouse
// WARN: call in the begining to avoid interupts
func GetLightHouse() *LigthHouse {
	once.Do(func() {
		ligthHouseInstance = newLH()
	})
	return ligthHouseInstance
}

func newLH() *LigthHouse {
	ioInstance := io.GetIO()
	elector := election.NewLeaderElection(ioInstance)
	ligthHouseInstance = &LigthHouse{io: ioInstance, elector: elector}
	return ligthHouseInstance
}

// ReadFollowers reads data from followers.
//
// Retrieves the data associated with all followers at the specified path.
// Returns the followers' data and any error encountered during reading.
func (t *LigthHouse) EnsurePathExists(path string) {
	if err := t.io.EnsurePathExists(path); err != nil {
		log.Fatal(err)
	}
}

// RegisterSequential writes `data` at `utils.Path`.
//
// Creates a sequential node at the specified path with the given data and returns the full path.
// Returns the sequential path where the data is written.
func (t *LigthHouse) RegisterSequential(path u.Path, data interface{}) u.Path {
	path, err := t.io.RegisterSequential(path, data)
	if err != nil {
		log.Fatal(err)
	}
	return path
}

// Read reads data at `path`.
//
// Retrieves the data stored at the specified path.
func (t *LigthHouse) Read(path u.Path) ([]byte, error) {
	return t.io.Read(path)
}

// Write writes data at `path`.
//
// Stores the provided data at the specified path.
func (t *LigthHouse) Write(path u.Path, newData any) {
	t.io.Write(path, newData)
}

func (t *LigthHouse) ReadAllChildsData(path u.Path) ([][]byte, error) {
	addrs, err := t.io.GetChildren(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read childs: %v", err)
	}
	var data [][]byte
	for _, addr := range addrs {
		fullPath := u.Path(u.PathBuilder{}.Base(path).CD(addr).Create())
		b, err := t.io.Read(fullPath)
		if err != nil {
			continue
		}
		data = append(data, b)
	}
	return data, nil
}
