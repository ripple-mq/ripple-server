package server

import (
	"context"

	"github.com/ripple-mq/ripple-server/internal/broker"
	"github.com/ripple-mq/ripple-server/internal/topic"
	"github.com/ripple-mq/ripple-server/pkg/utils"
	pb "github.com/ripple-mq/ripple-server/server/internal/proto"
)

type PCServer struct {
	Paddr string
	Caddr string
}

// CreateBucket creates a new bucket by initializing a new broker instance for the given topic and bucket.
// Starts Pub/Sub server & returns address
func (t Server) CreateBucket(ctx context.Context, req *pb.CreateBucketReq) (*pb.CreateBucketResp, error) {
	b := broker.NewBroker(topic.TopicBucket{TopicName: req.Topic, BucketName: req.Bucket})
	paddr, caddr := utils.RandLocalAddr(), utils.RandLocalAddr()
	if err := b.Run(paddr, caddr); err != nil {
		return &pb.CreateBucketResp{Success: false}, err
	}
	return &pb.CreateBucketResp{Success: true}, nil
}
