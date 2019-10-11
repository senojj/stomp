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

type request struct {
	Frame *proto.ClientFrame
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

func process(writer io.Writer, reader *proto.FrameReader) *processor {
	writeCh := make(chan request)
	readCh := make(chan packet)
	done := make(chan struct{}, 1)
	var receipts receiptMap
	ticker := time.NewTicker(time.Nanosecond)
	var wg sync.WaitGroup
	frameWriter := proto.NewFrameWriter(writer)

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
				fmt.Printf("read <- %s\n", frame)
				if nil != frame {
					switch frame.Command {
					case proto.CmdReceipt:
						id, ok := frame.Header.Get(proto.HdrReceiptId)

						if ok {
							rch, has := receipts.Get(id)

							if has {
								rch <- nil
								receipts.Del(id)
							}
						} else {
							log.Printf("error: stomp: unknown receipt-id: %s", id)
						}
					case proto.CmdError:
						id, ok := frame.Header.Get(proto.HdrReceiptId)
						content, rdErr := ioutil.ReadAll(frame.Body)

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
					case proto.CmdMessage:
						id, ok := frame.Header.Get(proto.HdrSubscription)

						if ok {
							var wg sync.WaitGroup

							message := Message{
								Header(frame.Header),
								&waitGroupReadCloser{
									reader: frame.Body,
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
					closeErr := frame.Body.Close()

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
				id, ok := wr.Frame.Header.Get(proto.HdrReceipt)

				if ok {
					receipts.Set(id, wr.C)
				}
				fmt.Printf("write -> %s\n", wr.Frame.String())
				_, wrErr := frameWriter.Write(wr.Frame)
				wr.C <- wrErr
			case <-done:
				done <- struct{}{}
				break loop
			}
		}
	}()

	return &processor{W: writeCh, R: readCh, done: done}
}
