package stomp

import "io"

type request struct {
	Frame *Frame
	Args  map[string]interface{}
}

type Session struct {
	Tx   chan<- request
	Rx   <-chan *Frame
	Err  <-chan error
	done chan struct{}
}

func NewSession(rw io.ReadWriter) *Session {
	err := make(chan error)
	trn := make(chan request)
	rcv := make(chan *Frame)
	done := make(chan struct{})

	receipts := make(map[string]chan<- struct{})

	go func() {
		transfer := newTx(rw, err)
		receive := newRx(rw, err)

	loop:
		for {
			select {
			case req := <-trn:
				id, hasReceipt := req.Frame.Header.Get(HdrReceipt)

				if hasReceipt {
					v, hasReceiptChan := req.Args["receipt"]

					if hasReceiptChan {
						ch, ok := v.(chan<- struct{})

						if ok {
							receipts[id] = ch
						}
					}
				}
				transfer.C <- req.Frame

			case frame := <-receive.C:
				switch frame.Command {
				case CmdReceipt:
					id, hasReceiptId := frame.Header.Get(HdrReceiptId)

					if hasReceiptId {
						ch, ok := receipts[id]

						if ok {
							ch <- struct{}{}
						}
					}
					closeErr := frame.Body.Close()

					if nil != closeErr {
						err <- closeErr
					}
				default:
					rcv <- frame
				}

			case <-done:
				break loop
			}
		}
		transfer.stop()
		receive.stop()
		close(trn)
		close(rcv)
		close(err)
	}()

	return &Session{
		Tx:   trn,
		Rx:   rcv,
		Err:  err,
		done: done,
	}
}
