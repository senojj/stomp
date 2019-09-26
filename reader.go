package stomp

import "io"

type measurableReader interface {
	io.Reader
	Len() int
}
