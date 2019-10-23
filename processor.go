package stomp

import (
	"fmt"
	"github.com/dynata/stomp/frame"
	"io"
	"io/ioutil"
	"log"
	"sync"
	"time"
)

type request struct {
	Frame *frame.Frame
	C     chan<- error
	Args  []interface{}
}

type packet struct {
	ID      string
	Message Message
}

type processor struct {
	W    chan<- request
	R    <-chan packet
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

func process(rw io.ReadWriter) *processor {
	writeCh := make(chan request)
	readCh := make(chan packet)
	done := make(chan struct{}, 1)
	var receipts receiptMap
	ticker := time.NewTicker(time.Nanosecond)
	var wg sync.WaitGroup

	go func() {
		wg.Add(1)
		defer wg.Done()

	loop:
		for {
			select {
			case <-ticker.C:
				f, rdErr := frame.Read(rw)

				if nil != rdErr {
					log.Println(rdErr)
					break loop
				}
				fmt.Printf("read <- %s\n", f)
				if nil != f {
					switch f.Command {
					case frame.CmdReceipt:
						id, ok := f.Header.Get(frame.HdrReceiptId)

						if ok {
							rch, has := receipts.Get(id)

							if has {
								rch <- nil
								receipts.Del(id)
							}
						} else {
							log.Printf("error: stomp: unknown receipt-id: %s", id)
						}
					case frame.CmdError:
						id, ok := f.Header.Get(frame.HdrReceiptId)
						content, rdErr := ioutil.ReadAll(f.Body)

						if nil != rdErr {
							log.Println(rdErr)
							continue
						}

						if ok {
							rch, has := receipts.Get(id)

							if has {
								rch <- fmt.Errorf(string(content))
								receipts.Del(id)
							}
						} else {
							log.Printf(string(content))
						}
					case frame.CmdMessage:
						id, ok := f.Header.Get(frame.HdrSubscription)

						if ok {
							var wg sync.WaitGroup

							message := Message{
								Header(f.Header),
								&waitGroupReadCloser{
									reader: f.Body,
									wg:     &wg,
								},
							}
							wg.Add(1)
							packet := packet{
								ID:      id,
								Message: message,
							}
							readCh <- packet
							wg.Wait()
						}
					}
					closeErr := f.Body.Close()

					if nil != closeErr {
						log.Printf("error: stomp: closing frame: %v", closeErr)
					}
				}
			case <-done:
				done <- struct{}{}
				ticker.Stop()
				close(readCh)
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
			case wr := <-writeCh:
				id, ok := wr.Frame.Header.Get(frame.HdrReceipt)

				if ok {
					receipts.Set(id, wr.C)
				}
				wrErr := wr.Frame.Write(rw)
				wr.C <- wrErr
			case <-done:
				done <- struct{}{}
				break loop
			}
		}
	}()

	return &processor{W: writeCh, R: readCh, done: done}
}
