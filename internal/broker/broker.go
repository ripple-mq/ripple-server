package broker

import (
	"github.com/charmbracelet/log"
	"github.com/ripple-mq/ripple-server/internal/broker/server"
	"github.com/ripple-mq/ripple-server/internal/lighthouse"
	lu "github.com/ripple-mq/ripple-server/internal/lighthouse/utils"
	tp "github.com/ripple-mq/ripple-server/internal/topic"
	"github.com/ripple-mq/ripple-server/pkg/utils"
)

type PCServerAddr struct {
	Paddr string
	Caddr string
}

type Broker struct {
	addr   PCServerAddr
	topic  tp.TopicBucket
	server *server.Server
}

func NewBroker(topic tp.TopicBucket) *Broker {
	paddr, caddr := utils.RandLocalAddr(), utils.RandLocalAddr()
	bs := server.NewServer(paddr, caddr)
	return &Broker{PCServerAddr{paddr, caddr}, topic, bs}
}

func (t *Broker) Run() error {
	if err := t.server.Listen(); err != nil {
		return err
	}
	if err := t.registerAndStartWatching(); err != nil {
		return err
	}
	return nil
}

// TODO: Avoid re-registering topic/bucket
// TODO: Cron job to push messages in batches to read replicas from leader
func (t *Broker) registerAndStartWatching() error {
	lh := lighthouse.GetLightHouse()
	path := t.topic.GetPath()

	followerPath, err := lh.RegisterAsFollower(path, t.addr)
	if err != nil {
		return err
	}
	fatalCh := lh.StartElectLoop(followerPath, t.addr, onBecommingLeader)
	go t.RunCleanupLoop(fatalCh)
	return nil
}

func (t *Broker) RunCleanupLoop(ch <-chan struct{}) {
	for range ch {
		t.server.Stop()
		return
	}
}

func onBecommingLeader(path lu.Path) {
	log.Infof("Heyyyyy, I became leader: %v", path)
}
