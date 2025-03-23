package server

import (
	"context"

	pb "github.com/ripple-mq/ripple-server/server/exposed/proto"
)

func (c Server) GetConsumerConnection(ctx context.Context, in *pb.GetConsumerConnnectionReq) (*pb.GetConsumerConnectionResp, error) {
	return nil, nil
}
