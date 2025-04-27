package server

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/log"
	"github.com/ripple-mq/ripple-server/internal/broker"
	"github.com/ripple-mq/ripple-server/internal/topic"
	pb "github.com/ripple-mq/ripple-server/server/exposed/proto"
	ib "github.com/ripple-mq/ripple-server/server/internal/proto"
	"google.golang.org/grpc"
)

func (c Server) CreateBucket(ctx context.Context, req *pb.CreateBucketReq) (*pb.CreateBucketResp, error) {
	tp := topic.TopicBucket{TopicName: req.Topic, BucketName: req.Bucket}
	b := broker.NewBroker(tp)
	servers, err := b.CreateBucket()
	if err != nil {
		return &pb.CreateBucketResp{Success: false}, err
	}

	fmt.Printf("servers: %s \n", servers)

	createReq(servers, tp)
	return nil, nil
}

func createReq(servers []broker.InternalRPCServerAddr, topic topic.TopicBucket) {

	for _, addr := range servers[:2] {
		func(addr string) {
			conn, err := grpc.NewClient(addr, grpc.WithInsecure())
			if err != nil {
				log.Errorf("did not connect: %v", err)
			}
			defer conn.Close()

			client := ib.NewInternalServiceClient(conn)
			client.CreateBucket(context.Background(), &ib.CreateBucketReq{Topic: topic.TopicName, Bucket: topic.BucketName})
		}(addr.Addr)

		time.Sleep(2 * time.Second)
	}

}
