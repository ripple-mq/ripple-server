package comm

import (
	"bytes"

	"github.com/ripple-mq/ripple-server/pkg/p2p/encoder"
)

// PCServerID holds Pub/Sub server addresses
type PCServerID struct {
	BrokerAddr string // Broker address
	ProducerID string // Producer ID
	ConsumerID string // Consumer ID
}

// DecodeToPCServerID decodes bytes to `PCServerID`
func DecodeToPCServerID(data []byte) (PCServerID, error) {
	var addr PCServerID
	err := encoder.GOBDecoder{}.Decode(bytes.NewBuffer(data), &addr)
	return addr, err
}
