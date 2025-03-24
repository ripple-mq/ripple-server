package server

import (
	"context"

	b "github.com/ripple-mq/ripple-server/internal/broker"
	"github.com/ripple-mq/ripple-server/internal/lighthouse"
	"github.com/ripple-mq/ripple-server/pkg/utils"
	pb "github.com/ripple-mq/ripple-server/server/internal/proto"
)

func (t Server) CreateBucket(context.Context, *pb.CreateBucketReq) (*pb.CreateBucketResp, error) {

	lh, _ := lighthouse.GetLightHouse()
	lh.Connect()

	broker := b.NewBroker()
	paddr, caddr := utils.RandLocalAddr(), utils.RandLocalAddr()
	broker.CreateAndRunQueue(paddr, caddr)

	return nil, nil
}
