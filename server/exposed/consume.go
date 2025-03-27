package server

import (
	"context"

	"github.com/ripple-mq/ripple-server/internal/broker"
	"github.com/ripple-mq/ripple-server/internal/broker/consumer"
	pb "github.com/ripple-mq/ripple-server/server/exposed/proto"
)

func (c Server) GetConsumerConnection(ctx context.Context, req *pb.GetConsumerConnnectionReq) (*pb.GetConsumerConnectionResp, error) {

	data, err := consumer.NewConsumer().GetServerConnection(req.Topic, req.Bucket)
	if err != nil {
		return &pb.GetConsumerConnectionResp{Success: false}, err
	}

	address, err := broker.DecodeToPCServerAddr(data)
	if err != nil {
		return &pb.GetConsumerConnectionResp{Success: false}, err
	}

	return &pb.GetConsumerConnectionResp{Address: address.Caddr}, nil
}
