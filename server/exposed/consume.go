package server

import (
	"context"

	"github.com/ripple-mq/ripple-server/internal/broker/comm"
	"github.com/ripple-mq/ripple-server/internal/broker/consumer"
	"github.com/ripple-mq/ripple-server/internal/topic"
	pb "github.com/ripple-mq/ripple-server/server/exposed/proto"
)

// GetConsumerConnection retrieves the consumer server connection details.
// Given a topic and bucket, it fetches the consumer server's connection address.
func (c Server) GetConsumerConnection(ctx context.Context, req *pb.GetConsumerConnnectionReq) (*pb.GetConsumerConnectionResp, error) {

	bucket := topic.TopicBucket{TopicName: req.Topic, BucketName: req.Bucket}
	data, err := consumer.NewConsumer(bucket).GetServerConnection(req.Topic, req.Bucket)
	if err != nil {
		return &pb.GetConsumerConnectionResp{Success: false}, err
	}

	cons, err := comm.DecodeToPCServerID(data)
	if err != nil {
		return &pb.GetConsumerConnectionResp{Success: false}, err
	}

	return &pb.GetConsumerConnectionResp{Address: cons.BrokerAddr, ConsumerId: cons.ConsumerID}, nil
}
