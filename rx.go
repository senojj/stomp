package stomp

import (
	"io"
	"sync"
	"time"
)

type rx struct {
	C    <-chan *Frame
	done chan struct{}
}

func newRx(r io.Reader, err chan<- error) *rx {
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
				var wg sync.WaitGroup

				wrc := waitingReadCloser{
					reader: f.Body,
					wg: &wg,
				}
				f.Body = &wrc
				ch <- f
				wg.Wait()
			case <-done:
				break loop
			}
		}
		ticker.Stop()
		close(ch)
	}()
	return &rx{C: ch, done: done}
}

func (x *rx) stop() {
	select {
	case x.done <- struct{}{}:
	default:
	}
}
