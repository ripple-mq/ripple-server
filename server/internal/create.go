package server

import (
	"context"

	"github.com/charmbracelet/log"
	"github.com/ripple-mq/ripple-server/internal/broker"
	"github.com/ripple-mq/ripple-server/internal/lighthouse"
	tp "github.com/ripple-mq/ripple-server/internal/topic"
	"github.com/ripple-mq/ripple-server/pkg/utils"
	pb "github.com/ripple-mq/ripple-server/server/internal/proto"
)

type PCServer struct {
	Paddr string
	Caddr string
}

func (t Server) CreateBucket(ctx context.Context, req *pb.CreateBucketReq) (*pb.CreateBucketResp, error) {

	b := broker.NewBroker()
	paddr, caddr := utils.RandLocalAddr(), utils.RandLocalAddr()
	b.CreateAndRunQueue(paddr, caddr)

	registerAndStartWatching(req.Topic, req.Bucket, paddr, caddr)

	return &pb.CreateBucketResp{Success: true}, nil
}

// TODO: Avoid re-registering topic/bucket
// TODO: Cron job to push messages in batches to read replicas from leader
func registerAndStartWatching(topic string, bucket string, paddr string, caddr string) {
	lh, _ := lighthouse.GetLightHouse()
	path := lh.RegisterFollower(tp.Topic{}.GetPath(topic, bucket), PCServer{Paddr: paddr, Caddr: caddr})

	if err := lh.ElectLeader(path, PCServer{Paddr: paddr, Caddr: caddr}); err != nil {
		log.Errorf("Got error while electing: %v, %v", err, path)
	}
	go lh.WatchForLeader(path, PCServer{Paddr: paddr, Caddr: caddr})
}
