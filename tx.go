package stomp

import "io"

type Tx struct {
	C    chan<- *Frame
	done chan struct{}
}

func newTx(w io.Writer, err chan<- error) *Tx {
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
	return &Tx{C: ch, done: done}
}

func (x *Tx) stop() {
	select {
	case x.done <- struct{}{}:
	default:
	}
}
