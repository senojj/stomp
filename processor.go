package stomp

import (
	"fmt"
	"github.com/dynata/stomp/proto"
	"io"
	"io/ioutil"
	"log"
	"sync"
	"time"
)

type writeRequest struct {
	Frame *proto.ClientFrame
	C     chan<- error
}

type processor struct {
	C    chan<- writeRequest
	done chan struct{}
	wg   sync.WaitGroup
}

func (p *processor) Close() error {
	select {
	case p.done <- struct{}{}:
	default:
	}
	p.wg.Wait()
	return nil
}

func process(writer io.Writer, reader *proto.FrameReader) *processor {
	ch := make(chan writeRequest)
	done := make(chan struct{}, 1)
	receipts := make(map[string]chan<- error)
	ticker := time.NewTicker(time.Nanosecond)
	var wg sync.WaitGroup

	go func() {
		wg.Add(1)
		defer wg.Done()

	loop:
		for {
			select {
			case <-ticker.C:
				frame, rdErr := reader.Read()

				if nil != rdErr {
					log.Println(rdErr)
					break loop
				}

				if nil != frame {
					switch frame.Command {
					case proto.CmdReceipt:
						id, ok := frame.Header.Get(proto.HdrReceiptId)

						if ok {
							receipts[id] <- nil
							delete(receipts, id)
						} else {
							log.Printf("error: stomp: unknown receipt-id: %s", id)
						}
						frame.Body.Close()
					case proto.CmdError:
						id, ok := frame.Header.Get(proto.HdrReceiptId)
						content, rdErr := ioutil.ReadAll(frame.Body)

						if nil != rdErr {
							log.Println(rdErr)
							continue
						}

						if ok {
							receipts[id] <- fmt.Errorf(string(content))
							delete(receipts, id)
						} else {
							log.Printf(string(content))
						}
						frame.Body.Close()
					}
				}
			case <-done:
				done <- struct{}{}
				ticker.Stop()
				break loop
			}
		}
	}()

	go func() {
		wg.Add(1)
		defer wg.Done()

	loop:
		for {
			select {
			case wr := <-ch:
				id, ok := wr.Frame.Header.Get(proto.HdrReceipt)

				if ok {
					receipts[id] = wr.C
				}
				_, wrErr := wr.Frame.WriteTo(writer)
				wr.C <- wrErr
			case <-done:
				done <- struct{}{}
				break loop
			}
		}
	}()

	return &processor{C: ch, done: done}
}
