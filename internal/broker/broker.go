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
	addr  PCServerAddr
	topic tp.TopicBucket
}

func NewBroker(topic tp.TopicBucket) *Broker {
	paddr, caddr := utils.RandLocalAddr(), utils.RandLocalAddr()
	return &Broker{PCServerAddr{paddr, caddr}, topic}
}

func (t *Broker) Run() {
	bs := server.NewServer()
	bs.Listen(t.addr.Paddr, t.addr.Caddr)
	t.registerAndStartWatching()
}

// TODO: Avoid re-registering topic/bucket
// TODO: Cron job to push messages in batches to read replicas from leader
func (t *Broker) registerAndStartWatching() {
	lh, _ := lighthouse.GetLightHouse()
	path := t.topic.GetPath()

	followerPath := lh.RegisterAsFollower(path, t.addr)
	lh.StartElectLoop(followerPath, t.addr, onBecommingLeader)
}

func onBecommingLeader(path lu.Path) {
	log.Infof("Heyyyyy, I became leader: %v", path)
}
