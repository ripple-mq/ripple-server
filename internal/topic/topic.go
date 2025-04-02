package topic

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/ripple-mq/ripple-server/internal/lighthouse/utils"
)

type TopicBucket struct {
	TopicName  string
	BucketName string
}

func (t TopicBucket) GetID() string {
	hash := md5.New()
	io.WriteString(hash, fmt.Sprintf("%s/%s", t.TopicName, t.BucketName))
	return hex.EncodeToString(hash.Sum(nil))
}

func (t TopicBucket) GetPath() utils.Path {
	return utils.PathBuilder{}.Base(utils.Root()).CD(t.TopicName).CD(t.BucketName).Create()
}
