package topic

import (
	"github.com/ripple-mq/ripple-server/internal/lighthouse/utils"
)

type TopicBucket struct {
	TopicName  string
	BucketName string
}

func (t TopicBucket) GetPath() utils.Path {
	return utils.PathBuilder{}.Base(utils.Root()).CD(t.TopicName).CD(t.BucketName).Create()
}
