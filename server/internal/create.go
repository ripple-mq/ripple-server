package server

import (
	"context"

	pb "github.com/ripple-mq/ripple-server/server/internal/proto"
)

func (t Server) CreateBucket(context.Context, *pb.CreateBucketReq) (*pb.CreateBucketResp, error) {
	return nil, nil
}
