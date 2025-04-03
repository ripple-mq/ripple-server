package server

import (
	"context"

	"github.com/ripple-mq/ripple-server/internal/broker/comm"
	"github.com/ripple-mq/ripple-server/internal/broker/producer"
	"github.com/ripple-mq/ripple-server/internal/topic"
	pb "github.com/ripple-mq/ripple-server/server/exposed/proto"
)

// GetProducerConnection retrieves the producer server connection details.
// Given a topic and bucket, it fetches the producer server's connection address.
func (c Server) GetProducerConnection(ctx context.Context, req *pb.GetProducerConnectionReq) (*pb.GetProducerConnectionResp, error) {

	bucket := topic.TopicBucket{TopicName: req.Topic, BucketName: req.Bucket}
	data, err := producer.NewProducer(bucket).GetServerConnection(req.Topic, req.Bucket)
	if err != nil {
		return &pb.GetProducerConnectionResp{Success: false}, err
	}

	prod, err := comm.DecodeToPCServerID(data)
	if err != nil {
		return &pb.GetProducerConnectionResp{Success: false}, err
	}

	return &pb.GetProducerConnectionResp{Address: prod.BrokerAddr, ProducerId: prod.ProducerID}, nil
}
