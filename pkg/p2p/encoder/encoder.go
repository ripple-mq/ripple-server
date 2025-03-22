package encoder

import "io"

type Encoder interface {
	Encode(value any, buff io.Writer) error
}
type Decoder interface {
	Decode(value io.Reader, res any) error
}
