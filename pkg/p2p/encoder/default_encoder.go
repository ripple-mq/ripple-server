package encoder

import (
	"encoding/gob"
	"io"
)

type GOBEncoder struct {
}
type GOBDecoder struct {
}

func (g GOBEncoder) Encode(value any, buff io.Writer) error {
	enc := gob.NewEncoder(buff)
	err := enc.Encode(value)
	return err
}

func (g GOBDecoder) Decode(value io.Reader, res any) error {
	dec := gob.NewDecoder(value)
	return dec.Decode(res)
}
