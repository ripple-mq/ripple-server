package election

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/ripple-mq/ripple-server/internal/lighthouse/io"
	u "github.com/ripple-mq/ripple-server/internal/lighthouse/utils"
)

type Role string

const (
	Leader   Role = "leader"
	Follower Role = "followers"
)

type LeaderElection struct {
	leaderSignal chan bool
	fatalSignal  chan struct{}
	io           *io.IO
}

// NewLeaderElection returns *LeaderElection
func NewLeaderElection(io *io.IO) *LeaderElection {
	return &LeaderElection{io: io, leaderSignal: make(chan bool, 1), fatalSignal: make(chan struct{}, 1)}
}

// Start begins the leader election process.
//
//   - It initiates the leader election by calling `elect`
//   - logs an error if it fails
//   - triggers the fatal signal.
//   - starts watching the leader election process asynchronously.
func (t *LeaderElection) Start(fpath u.Path, data any) {
	if err := t.elect(fpath, data); err != nil {
		log.Errorf("failed to elect leader: %v", err)
		t.FatalSignal()
	}
	go t.watch(fpath, data)
}

// elect attempts to become the leader by selecting the smallest follower.
//
// It checks the followers' list, sorts them, and if the current node is the smallest,
// it registers itself as the leader and signals leadership. Otherwise, it logs the current leader.
func (t *LeaderElection) elect(fpath u.Path, data any) error {

	followersPath := u.PathBuilder{}.Base(fpath).CDBack().Create()
	followers, err := t.io.GetChildren(followersPath)
	if err != nil {
		return fmt.Errorf("failed to get childrens of %s, %v", followersPath, err)
	}
	if len(followers) == 0 {
		return fmt.Errorf("no followers found at %s, %v", followersPath, err)
	}

	sort.Slice(followers, func(i, j int) bool {
		ni, _ := strconv.Atoi(strings.Split(followers[i], "-")[1])
		nj, _ := strconv.Atoi(strings.Split(followers[j], "-")[1])
		return ni < nj
	})

	fileName := u.PathBuilder{}.Base(fpath).FileName()
	if followers[0] == fileName {
		if _, err := t.RegisterLeader(u.PathBuilder{}.Base(fpath).CDBack().CDBack().Create(), data); err != nil {
			return err
		}
		t.LeaderSignal(true)
	} else {
		log.Infof("I am not the leader, my node is: %s\n  but leader: %s", fileName, followers[0])
		t.LeaderSignal(false)
	}

	return nil
}

// watch monitors changes in the leader's status and re-elects if needed.
//
// It listens for changes in the leader's path and triggers a re-election process
// whenever the leader changes.
func (t *LeaderElection) watch(fpath u.Path, data any) {
	defer t.FatalSignal()

	for {
		leaderPath := u.PathBuilder{}.Base(fpath).CDBack().CDBack().CD(string(Leader)).Create()
		children, ch, err := t.io.GetChildrenAndWatch(leaderPath)
		if err != nil {
			log.Errorf("failed to get childrens of %v, %v", leaderPath, err)
			return
		}
		if len(children) > 0 {
			leader := children[0]
			log.Debugf("Current Leader: %s\n", leader)
		}

		<-ch
		log.Debugf("Leader path changed, re-electing leader.")
		err = t.elect(fpath, data)
		if err != nil {
			log.Errorf("failed to watch for leader: %v, %v", err, leaderPath)
			return
		}
	}
}

// RegisterFollower registers the current node as a follower at the given path.
//
// It creates a new path for the follower, registers it sequentially,
// and returns the path if successful, or an error if the registration fails.
func (t *LeaderElection) RegisterFollower(path u.Path, data any) (u.Path, error) {
	path = u.PathBuilder{}.Base(path).CD(string(Follower)).Create()
	path, err := t.io.RegisterSequential(path, data)
	if err != nil {
		return u.Path{}, fmt.Errorf("failed to register as follower: %v", err)
	}
	return path, nil
}

// RegisterLeader registers the current node as the leader at the given path.
//
// It creates a new path for the leader, registers it sequentially,
// and returns the path if successful, or an error if the registration fails.
func (t *LeaderElection) RegisterLeader(path u.Path, data any) (u.Path, error) {
	path = u.PathBuilder{}.Base(path).CD(string(Leader)).Create()
	path, err := t.io.RegisterSequential(path, data)
	if err != nil {
		return u.Path{}, fmt.Errorf("failed to register as leader: %v", err)
	}
	return path, nil
}

// ReadLeader retrieves the leader's data from the specified path.
//
// It checks the leader directory, fetches the leader's child node, and reads its data.
// Returns the leader's data or an error if no leader is found or reading fails.
func (t *LeaderElection) ReadLeader(path u.Path) ([]byte, error) {
	leaderDir := u.PathBuilder{}.Base(path).CD(string(Leader)).Create()
	childs, err := t.io.GetChildren(leaderDir)
	if err != nil || len(childs) == 0 {
		return nil, fmt.Errorf("no leader found: %v", err)
	}

	leaderPath := u.Path(u.PathBuilder{}.Base(leaderDir).CD(childs[0]).Create())
	data, err := t.io.Read(leaderPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read leader: %v", err)
	}
	return data, nil
}

// ReadFollowers retrieves the data of all followers from the specified path.
//
// It checks the followers directory, fetches each follower's data, and returns a slice
// containing the data of all followers. If no followers are found or reading fails, an error is returned
func (t *LeaderElection) ReadFollowers(path u.Path) ([][]byte, error) {
	followersDir := u.PathBuilder{}.Base(path).CD(string(Follower)).Create()
	childs, err := t.io.GetChildren(followersDir)
	if err != nil || len(childs) == 0 {
		return nil, fmt.Errorf("no followers found: %v", err)
	}

	followers := [][]byte{}

	for _, child := range childs {
		followerPath := u.Path(u.PathBuilder{}.Base(followersDir).CD(child).Create())
		if data, err := t.io.Read(followerPath); err == nil {
			followers = append(followers, data)
		}
	}
	if len(followers) == 0 {
		return nil, fmt.Errorf("failed to read follower: %v", err)
	}
	return followers, nil
}

// ListenForLeaderSignal returns a channel that signals when the leader is elected.
func (t *LeaderElection) ListenForLeaderSignal() <-chan bool {
	return t.leaderSignal
}

// LeaderSignal sends a signal indicating that the leader has been elected.
func (t *LeaderElection) LeaderSignal(val bool) {
	t.leaderSignal <- val
}

// FatalSignal sends a signal indicating a fatal error has occurred.
func (t *LeaderElection) FatalSignal() {
	t.fatalSignal <- struct{}{}
}

// ListenForFatalSignal returns a channel that signals when a fatal error occurs.
func (t *LeaderElection) ListenForFatalSignal() <-chan struct{} {
	return t.fatalSignal
}
