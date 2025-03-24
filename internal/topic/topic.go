package topic

import (
	"github.com/ripple-mq/ripple-server/internal/lighthouse/utils"
)

type Topic struct {
}

func (t Topic) GetPath(topic string, bucket string) utils.Path {
	return utils.PathBuilder{}.Base(utils.Root()).CD(topic).CD(bucket).Create()
}
