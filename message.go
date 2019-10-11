package stomp

import (
	"github.com/dynata/stomp/proto"
	"io"
)

type Header proto.Header

func (h Header) Get(name string) (string, bool) {
	return proto.Header(h).Get(name)
}

type Message struct {
	Header       Header
	Body         io.ReadCloser
}
