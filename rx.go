package stomp

import (
	"io"
	"time"
)

type Rx struct {
	C    <-chan *Frame
	done chan struct{}
}

func newRx(r io.Reader, err chan<- error) *Rx {
	ch := make(chan *Frame)
	done := make(chan struct{}, 1)

	go func() {
		ticker := time.NewTicker(time.Nanosecond)

	loop:
		for {
			select {
			case <-ticker.C:
				f, readErr := ReadFrame(r)

				if nil != readErr {
					err <- readErr
				}
				wrc := newWaitingReadCloser(f.Body)
				f.Body = wrc
				ch <- f
				wrc.wait()
			case <-done:
				break loop
			}
		}
		ticker.Stop()
		close(ch)
	}()
	return &Rx{C: ch, done: done}
}

func (x *Rx) stop() {
	select {
	case x.done <- struct{}{}:
	default:
	}
}
