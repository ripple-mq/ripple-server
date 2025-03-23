package lighthouse

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"
)

type Role string

const (
	Leader   Role = "leader"
	Follower Role = "follower"
)

type Path struct {
	Base string
	Role Role
	Name string
}

func (t *Path) AsLeader() string {
	return fmt.Sprintf("%s/%s", t.Base, Leader)
}

func (t *Path) AsFollower() string {
	return fmt.Sprintf("%s/%s", t.Base, Follower)
}

func (t *Path) SetFollower() *Path {
	t.Role = Follower
	return t
}

func (t *Path) FullPath() string {
	path := t.BasePath()
	if t.Name != "" {
		path += "/" + t.Name
	}
	return path
}

func (t *Path) BasePath() string {
	path := t.Base
	if t.Role != "" {
		path += "/" + string(t.Role)
	}
	return path
}

func (t *LigthHouse) RegisterFollower(path Path, data string) Path {
	return t.RegisterSequential(Path{Base: path.Base, Role: Follower}, data)
}

func (t *LigthHouse) RegisterLeader(path Path, data string) Path {
	return t.RegisterSequential(Path{Base: path.Base, Role: Leader, Name: path.Name}, data)
}

func (t *LigthHouse) ElectLeader(path Path, data string) error {
	children, _, err := t.conn.Children(path.BasePath())
	if err != nil {
		return err
	}

	sort.Slice(children, func(i, j int) bool {
		ni, _ := strconv.Atoi(strings.Split(children[i], "-")[1])
		nj, _ := strconv.Atoi(strings.Split(children[j], "-")[1])
		return ni < nj
	})

	if path.FullPath() == path.BasePath()+"/"+children[0] {
		t.RegisterSequential(Path{Base: path.Base, Role: Leader, Name: path.Name}, data)
		fmt.Printf("I am the leader: %s\n", path.FullPath())
	} else {
		fmt.Printf("I am not the leader, my node is: %s\n  but leader: %s", path.FullPath(), path.BasePath()+"/"+children[0])
	}

	return nil
}

func (t *LigthHouse) WatchForLeader(path Path, data string) error {
	for {
		log.Info(path.AsLeader())
		children, _, ch, err := t.conn.ChildrenW(path.AsLeader())
		if err != nil {
			return err
		}
		if len(children) > 0 {
			leader := children[0]
			fmt.Printf("Current Leader: %s\n", leader)
		}

		<-ch
		fmt.Println("Leader path changed, re-electing leader.")
		err = t.ElectLeader(path, data)
		if err != nil {
			return err
		}
	}
}
