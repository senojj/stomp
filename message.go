package stomp

import (
	"github.com/dynata/stomp/proto"
	"io"
)

type Header proto.Header

type Message struct {
	Header Header
	Body io.ReadCloser
}