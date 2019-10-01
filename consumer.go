package stomp

import (
	"github.com/dynata/stomp/proto"
	"log"
	"sync"
)

type readResult struct {
	Frame *proto.ServerFrame
	Err   error
}

type consumer struct {
	C    <-chan readResult
	done chan struct{}
	wg   sync.WaitGroup
}

func (i *consumer) Close() error {
	select {
	case i.done <- struct{}{}:
	default:
	}
	i.wg.Wait()
	return nil
}

func consume(r *proto.FrameReader) *consumer {
	ch := make(chan readResult)
	done := make(chan struct{}, 1)
	var wg sync.WaitGroup

	go func() {
		wg.Add(1)
		defer wg.Done()

	loop:
		for {
			frame, err := r.Read()

			if nil != frame {
				log.Printf(frame.String())
			} else {
				log.Printf("<nil>")
			}
			select {
			case ch <- readResult{frame, err}:
			case <-done:
				break loop
			}
		}
	}()
	return &consumer{C: ch, done: done, wg: wg}
}