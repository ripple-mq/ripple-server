package lighthouse

import (
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"
	u "github.com/ripple-mq/ripple-server/internal/lighthouse/utils"
)

type Role string

const (
	Leader   Role = "leader"
	Follower Role = "followers"
)

func (t *LigthHouse) RegisterFollower(path u.Path, data any) u.Path {
	path = u.PathBuilder{}.Base(path).CD(string(Follower)).Create()
	return t.RegisterSequential(path, data)
}

func (t *LigthHouse) RegisterLeader(path u.Path, data any) u.Path {
	path = u.PathBuilder{}.Base(path).CD(string(Leader)).Create()
	return t.RegisterSequential(path, data)
}

func (t *LigthHouse) ElectLeader(fpath u.Path, data any) error {

	leaderPath := u.PathBuilder{}.Base(fpath).CDBack().GetFile()
	children, _, err := t.conn.Children(leaderPath)
	if err != nil || len(children) == 0 {
		log.Errorf("ElectLeader() %s %s", leaderPath, fpath)
		return err
	}

	sort.Slice(children, func(i, j int) bool {
		ni, _ := strconv.Atoi(strings.Split(children[i], "-")[1])
		nj, _ := strconv.Atoi(strings.Split(children[j], "-")[1])
		return ni < nj
	})

	fileName := u.PathBuilder{}.Base(fpath).FileName()
	if children[0] == fileName {
		t.RegisterLeader(u.PathBuilder{}.Base(fpath).CDBack().CDBack().Create(), data)
		log.Infof("I am the leader: %s\n", fpath)
	} else {
		log.Infof("I am not the leader, my node is: %s\n  but leader: %s", fileName, children[0])
	}

	return nil
}

func (t *LigthHouse) WatchForLeader(fpath u.Path, data any) {
	for {
		leaderPath := u.PathBuilder{}.Base(fpath).CDBack().CDBack().CD(string(Leader)).GetFile()
		children, _, ch, err := t.conn.ChildrenW(leaderPath)
		if err != nil {
			log.Fatalf("failed to watch for leader: %v %s", err, leaderPath)
		}
		if len(children) > 0 {
			leader := children[0]
			log.Debugf("Current Leader: %s\n", leader)
		}

		<-ch
		log.Debugf("Leader path changed, re-electing leader.")
		err = t.ElectLeader(fpath, data)
		if err != nil {
			log.Errorf("failed to watch for leader: %v, %v", err, leaderPath)
			return
		}
	}
}
