package stomp

import "io"

type txrx struct {
	Tx   chan<- *Frame
	Rx   <-chan *Frame
	Err  <-chan error
	done <-chan struct{}
}

func process(rw io.ReadWriter) *txrx {

}
