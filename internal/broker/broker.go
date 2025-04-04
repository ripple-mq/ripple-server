package broker

import (
	"bytes"
	"fmt"

	"github.com/ripple-mq/ripple-server/internal/broker/comm"
	"github.com/ripple-mq/ripple-server/internal/broker/consumer/loadbalancer"
	"github.com/ripple-mq/ripple-server/internal/broker/server"
	"github.com/ripple-mq/ripple-server/internal/lighthouse"
	lu "github.com/ripple-mq/ripple-server/internal/lighthouse/utils"
	tp "github.com/ripple-mq/ripple-server/internal/topic"
	"github.com/ripple-mq/ripple-server/pkg/p2p/encoder"
	"github.com/ripple-mq/ripple-server/pkg/utils/config"
	"github.com/ripple-mq/ripple-server/pkg/utils/env"
)

type InternalRPCServerAddr struct {
	Addr string
}

type Broker struct {
	topic         tp.TopicBucket
	WatchLeaderCh chan struct{}
}

// NewBroker returns `*Broker` with specified topic
func NewBroker(topic tp.TopicBucket) *Broker {
	return &Broker{topic: topic, WatchLeaderCh: make(chan struct{})}
}

// Run spins up Pub/Sub servers & starts listening to new conn
func (t *Broker) Run(pId, cId string) error {
	bs := server.NewServer(pId, cId, t.topic)
	if err := bs.Listen(); err != nil {
		return err
	}
	addr := fmt.Sprintf("%s:%d", env.Get("ZK_IPv4", ""), config.Conf.AsyncTCP.Port)
	if err := t.registerAndStartWatching(bs, comm.PCServerID{BrokerAddr: addr, ProducerID: pId, ConsumerID: cId}); err != nil {
		return err
	}
	return nil
}

// registerAndStartWatching registers broker as follower & watches leader
//
// TODO: Avoid re-registering topic/bucket
// TODO: Cron job to push messages in batches to read replicas from leader
func (t *Broker) registerAndStartWatching(bs *server.Server, addr comm.PCServerID) error {
	lh := lighthouse.GetLightHouse()
	path := t.topic.GetPath()

	followerPath, err := lh.RegisterAsFollower(path, addr)
	if err != nil {
		return err
	}
	fatalCh := lh.StartElectLoop(followerPath, addr, t.WatchLeaderCh)
	go t.watchLeader(bs)
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

// watchLeader will be executed when current broker becomes leader
//
// TODO: Spin up cron job to distribute messages to follower in batches
func (t *Broker) watchLeader(bs *server.Server) {
	for {
		<-t.WatchLeaderCh
		bs.InformLeaderStatus()
	}
}

func (t *Broker) CreateBucket() ([]InternalRPCServerAddr, error) {
	servers, err := t.getAllServers()
	if err != nil {
		return nil, err
	}
	nodes := []InternalRPCServerAddr{}
	lb := loadbalancer.NewReadReqLoadBalancer()
	for range config.Conf.Topic.Replicas {
		var node InternalRPCServerAddr
		serverNode := servers[lb.GetIndex(len(servers)-1)]
		err := encoder.GOBDecoder{}.Decode(bytes.NewBuffer(serverNode), &node)
		if err != nil {
			continue
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

func (t *Broker) getAllServers() ([][]byte, error) {
	lh := lighthouse.GetLightHouse()
	data, err := lh.ReadAllChildsData(lu.PathBuilder{}.Base(lu.Root()).CD("servers").Create())
	if err != nil || len(data) == 0 {
		return nil, err
	}
	return data, nil
}
