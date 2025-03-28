package broker

import (
	"bytes"

	"github.com/charmbracelet/log"
	"github.com/ripple-mq/ripple-server/internal/broker/server"
	"github.com/ripple-mq/ripple-server/internal/lighthouse"
	lu "github.com/ripple-mq/ripple-server/internal/lighthouse/utils"
	tp "github.com/ripple-mq/ripple-server/internal/topic"
	"github.com/ripple-mq/ripple-server/pkg/p2p/encoder"
)

// PCServerAddr holds Pub/Sub server addresses
type PCServerAddr struct {
	Paddr string // Producer address
	Caddr string // Consumer address
}

// DecodeToPCServerAddr decodes bytes to `PCServerAddr`
func DecodeToPCServerAddr(data []byte) (PCServerAddr, error) {
	var addr PCServerAddr
	err := encoder.GOBDecoder{}.Decode(bytes.NewBuffer(data), &addr)
	return addr, err
}

type Broker struct {
	topic tp.TopicBucket
}

// NewBroker returns `*Broker` with specified topic
//
// Returns:
//   - *Broker
func NewBroker(topic tp.TopicBucket) *Broker {
	return &Broker{topic}
}

// Run spins up Pub/Sub servers & starts listening to new conn
//
// Returns:
//   - error
func (t *Broker) Run(paddr, caddr string) error {
	bs := server.NewServer(paddr, caddr)
	if err := bs.Listen(); err != nil {
		return err
	}
	if err := t.registerAndStartWatching(bs, PCServerAddr{Paddr: paddr, Caddr: caddr}); err != nil {
		return err
	}
	return nil
}

// registerAndStartWatching registers broker as follower & watches leader
//
// Returns:
//   - error
//
// TODO: Avoid re-registering topic/bucket
// TODO: Cron job to push messages in batches to read replicas from leader
func (t *Broker) registerAndStartWatching(bs *server.Server, addr PCServerAddr) error {
	lh := lighthouse.GetLightHouse()
	path := t.topic.GetPath()

	followerPath, err := lh.RegisterAsFollower(path, addr)
	if err != nil {
		return err
	}
	fatalCh := lh.StartElectLoop(followerPath, addr, onBecommingLeader)
	go t.RunCleanupLoop(bs, fatalCh)
	return nil
}

// RunCleanupLoop gracefully stutdowns Pub/Sub server
//
//	Async
func (t *Broker) RunCleanupLoop(server *server.Server, ch <-chan struct{}) {
	for range ch {
		server.Stop()
		return
	}
}

// onBecommingLeader will be executed when current broker becomes leader
//
// TODO: Spin up cron job to distribute messages to follower in batches
func onBecommingLeader(path lu.Path) {
	log.Infof("Heyyyyy, I became leader: %v", path)
}
