package server

import (
	"context"

	pb "github.com/ripple-mq/ripple-server/server/exposed/proto"
)

func (c Server) CreateBucket(ctx context.Context, in *pb.CreateBucketReq) (*pb.CreateBucketResp, error) {
	return nil, nil
}
