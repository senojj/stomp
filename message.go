package stomp

import (
	"github.com/dynata/stomp/frame"
	"io"
)

type Header frame.Header

func (h Header) Get(name string) (string, bool) {
	return frame.Header(h).Get(name)
}

type Message struct {
	Header       Header
	Body         io.ReadCloser
}
