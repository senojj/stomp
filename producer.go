package stomp

import (
	"github.com/dynata/stomp/proto"
	"io"
	"sync"
)

type writeRequest struct {
	Frame *proto.ClientFrame
	C     chan<- error
}

type writeResult struct {
	Written int64
	Err     error
}

type producer struct {
	C    chan<- writeRequest
	done chan struct{}
	wg   sync.WaitGroup
}

func (i *producer) Close() error {
	select {
	case i.done <- struct{}{}:
	default:
	}
	i.wg.Wait()
	return nil
}

func produce(w io.Writer) *producer {
	ch := make(chan writeRequest)
	done := make(chan struct{}, 1)
	var wg sync.WaitGroup

	go func() {
		wg.Add(1)
		defer wg.Done()

	loop:
		for {
			select {
			case req, ok := <-ch:
				if !ok {
					break loop
				}
				_, err := req.Frame.WriteTo(w)
				req.C <- err
			case <-done:
				break loop
			}
		}
	}()
	return &producer{C: ch, done: done, wg: wg}
}
