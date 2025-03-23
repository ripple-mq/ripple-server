package utils

import (
	"bytes"
	"encoding/gob"
)

func ToBytes(data any) []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	_ = enc.Encode(data)
	return buf.Bytes()
}
