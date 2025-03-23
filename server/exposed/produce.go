package server

import (
	"context"

	pb "github.com/ripple-mq/ripple-server/server/exposed/proto"
)

func (c Server) GetProducerConnection(ctx context.Context, in *pb.GetProducerConnectionReq) (*pb.GetProducerConnectionResp, error) {
	return nil, nil
}
