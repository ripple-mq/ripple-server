package server

import (
	"context"

	"github.com/ripple-mq/ripple-server/internal/broker"
	"github.com/ripple-mq/ripple-server/internal/topic"
	pb "github.com/ripple-mq/ripple-server/server/internal/proto"
)

type PCServer struct {
	Paddr string
	Caddr string
}

func (t Server) CreateBucket(ctx context.Context, req *pb.CreateBucketReq) (*pb.CreateBucketResp, error) {
	b := broker.NewBroker(topic.TopicBucket{TopicName: req.Topic, BucketName: req.Bucket})
	b.Run()
	return &pb.CreateBucketResp{Success: true}, nil
}
