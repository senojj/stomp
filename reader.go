package stomp

import "io"

type MeasurableReader interface {
	io.Reader
	Len() int
}
