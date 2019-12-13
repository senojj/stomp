package stomp

import "io"

type tx struct {
	C    chan<- *Frame
	done chan struct{}
}

func newTx(w io.Writer, err chan<- error) *tx {
	ch := make(chan *Frame)
	done := make(chan struct{}, 1)

	go func() {
	loop:
		for {
			select {
			case f := <-ch:
				writeErr := f.Write(w)

				if nil != writeErr {
					err <- writeErr
				}
			case <-done:
				break loop
			}
		}
		close(ch)
	}()
	return &tx{C: ch, done: done}
}

func (x *tx) stop() {
	select {
	case x.done <- struct{}{}:
	default:
	}
}
