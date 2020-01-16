package stomp

import "io"

type Session struct {
	Tx   *Tx
	Rx   *Rx
	Err  <-chan error
	done chan struct{}
}

func NewSession(rw io.ReadWriter) *Session {
		chError := make(chan error)
		transfer := newTx(rw, chError)
		receive := newRx(rw, chError)
		chDone := make(chan struct{}, 1)

		return &Session{
			Tx: transfer,
			Rx: receive,
			Err: chError,
			done: chDone,
		}
}
