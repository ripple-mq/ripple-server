package broker

import (
	"bytes"

	"github.com/charmbracelet/log"
	"github.com/ripple-mq/ripple-server/internal/broker/consumer/loadbalancer"
	"github.com/ripple-mq/ripple-server/internal/broker/server"
	"github.com/ripple-mq/ripple-server/internal/lighthouse"
	"github.com/ripple-mq/ripple-server/internal/lighthouse/utils"
	lu "github.com/ripple-mq/ripple-server/internal/lighthouse/utils"
	tp "github.com/ripple-mq/ripple-server/internal/topic"
	"github.com/ripple-mq/ripple-server/pkg/p2p/encoder"
	"github.com/ripple-mq/ripple-server/pkg/utils/config"
)

type InternalRPCServerAddr struct {
	Addr string
}

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
func NewBroker(topic tp.TopicBucket) *Broker {
	return &Broker{topic}
}

// Run spins up Pub/Sub servers & starts listening to new conn
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

// RunCleanupLoop gracefully shuts down the Pub/Sub server when signaled.
//
// It listens for a signal on the provided channel and stops the server when
// the signal is received, ensuring a clean shutdown.
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

//  /servers
//		/s-0000000001010101202nsa02 internal  gRPC address

func (t *Broker) CreateBucket() ([]InternalRPCServerAddr, error) {
	servers, err := t.getAllServers()
	if err != nil {
		return nil, err
	}
	nodes := []InternalRPCServerAddr{}
	lb := loadbalancer.NewReadReqLoadBalancer()
	for range config.Conf.Topic.Replicas {
		var node InternalRPCServerAddr

		encoder.GOBDecoder{}.Decode(bytes.NewBuffer(servers[lb.GetIndex(len(servers))]), node)
		nodes = append(nodes, node)
	}
	return nodes, nil
}

func (t *Broker) getAllServers() ([][]byte, error) {
	lh := lighthouse.GetLightHouse()
	data, err := lh.ReadAllChildsData(utils.PathBuilder{}.Base(utils.Root()).CD("servers").Create())
	if err != nil || len(data) == 0 {
		return nil, err
	}
	return data, nil
}
