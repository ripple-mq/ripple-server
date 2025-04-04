package server

import (
	"context"

	"github.com/ripple-mq/ripple-server/internal/broker"
	"github.com/ripple-mq/ripple-server/internal/broker/producer"
	"github.com/ripple-mq/ripple-server/pkg/utils/config"
	pb "github.com/ripple-mq/ripple-server/server/exposed/proto"
)

// GetProducerConnection retrieves the producer server connection details.
// Given a topic and bucket, it fetches the producer server's connection address.
func (c Server) GetProducerConnection(ctx context.Context, req *pb.GetProducerConnectionReq) (*pb.GetProducerConnectionResp, error) {

	data, err := producer.NewProducer().GetServerConnection(req.Topic, req.Bucket)
	if err != nil {
		return &pb.GetProducerConnectionResp{Success: false}, err
	}

	prod, err := broker.DecodeToPCServerID(data)
	if err != nil {
		return &pb.GetProducerConnectionResp{Success: false}, err
	}

	return &pb.GetProducerConnectionResp{Address: config.Conf.AsyncTCP.Addr, ProducerId: prod.ProducerID}, nil
}
