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
	leaderSignal chan struct{}
	fatalSignal  chan struct{}
	io           *io.IO
}

func NewLeaderElection(io *io.IO) *LeaderElection {
	return &LeaderElection{io: io, leaderSignal: make(chan struct{}, 1), fatalSignal: make(chan struct{}, 1)}
}

func (t *LeaderElection) Start(fpath u.Path, data any) {
	if err := t.elect(fpath, data); err != nil {
		log.Errorf("failed to elect leader: %v", err)
		t.FatalSignal()
	}
	go t.watch(fpath, data)
}

func (t *LeaderElection) elect(fpath u.Path, data any) error {

	followersPath := u.PathBuilder{}.Base(fpath).CDBack().GetFile()
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
		t.LeaderSignal()
	} else {
		log.Infof("I am not the leader, my node is: %s\n  but leader: %s", fileName, followers[0])
	}

	return nil
}

func (t *LeaderElection) watch(fpath u.Path, data any) {
	defer t.FatalSignal()

	for {
		leaderPath := u.PathBuilder{}.Base(fpath).CDBack().CDBack().CD(string(Leader)).GetFile()
		children, ch, err := t.io.GetChildrenAndWatch(leaderPath)
		if err != nil {
			log.Errorf("failed to get childrens of %s, %v", leaderPath, err)
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

func (t *LeaderElection) RegisterFollower(path u.Path, data any) (u.Path, error) {
	path = u.PathBuilder{}.Base(path).CD(string(Follower)).Create()
	path, err := t.io.RegisterSequential(path, data)
	if err != nil {
		return u.Path{}, fmt.Errorf("failed to register as follower: %v", err)
	}
	return path, nil
}

func (t *LeaderElection) RegisterLeader(path u.Path, data any) (u.Path, error) {
	path = u.PathBuilder{}.Base(path).CD(string(Leader)).Create()
	path, err := t.io.RegisterSequential(path, data)
	if err != nil {
		return u.Path{}, fmt.Errorf("failed to register as leader: %v", err)
	}
	return path, nil
}

func (t *LeaderElection) ListenForLeaderSignal() <-chan struct{} {
	return t.leaderSignal
}

func (t *LeaderElection) LeaderSignal() {
	t.leaderSignal <- struct{}{}
}

func (t *LeaderElection) ListenForFatalSignal() <-chan struct{} {
	return t.fatalSignal
}

func (t *LeaderElection) FatalSignal() {
	t.fatalSignal <- struct{}{}
}
