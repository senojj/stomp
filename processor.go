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
	Args  []interface{}
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
	var receipts receiptMap
	var subscriptions subscriptionMap
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
						id, ok := frame.Header.Get(proto.HdrId)

						if ok {
							fn, has := subscriptions.Get(id)

							if has {
								message := Message{
									Header(frame.Header),
									frame.Body,
								}
								fn(message)
							}
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
				if wr.Frame.Command == proto.CmdSubscribe {
					id, ok := wr.Frame.Header.Get(proto.HdrId)

					if ok {
						if len(wr.Args) > 0 {
							v, ok := wr.Args[0].(func(Message))

							if ok {
								subscriptions.Set(id, v)
							}
						}
					}
				}
				id, ok := wr.Frame.Header.Get(proto.HdrReceipt)

				if ok {
					receipts.Set(id, wr.C)
				}
				_, wrErr := frameWriter.Write(wr.Frame)
				wr.C <- wrErr
			case <-done:
				done <- struct{}{}
				break loop
			}
		}
	}()

	return &processor{C: ch, done: done}
}
