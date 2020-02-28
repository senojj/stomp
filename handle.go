package stomp

import (
	"context"
	"io"
	"time"
)

var bytesNewLine = []byte{byteNewLine}

type rxpkg struct {
	frame *Frame
	err   error
}

type rx struct {
	c    <-chan rxpkg
	done chan struct{}
}

func newRx(r io.Reader) rx {
	ch := make(chan rxpkg)
	done := make(chan struct{}, 1)

	go func() {
		ticker := time.NewTicker(time.Nanosecond)

	loop:
		for {
			select {
			case <-ticker.C:
				f, readErr := ReadFrame(r)

				if nil != readErr {
					ch <- rxpkg{nil, readErr}
				}

				if nil != f {
					wrc := newWaitingReadCloser(f.Body)
					f.Body = wrc
					ch <- rxpkg{f, nil}
					wrc.wait()
				} else {
					ch <- rxpkg{nil, nil}
				}
			case <-done:
				break loop
			}
		}
		ticker.Stop()
		ch <- rxpkg{nil, io.EOF}
		close(ch)
	}()

	return rx{ch, done}
}

func (x rx) stop() {
	select {
	case x.done <- struct{}{}:
	default:
	}
}

type txpkg struct {
	frame *Frame
	err   chan<- error
}

type tx struct {
	c    chan<- txpkg
	done chan struct{}
}

func newTx(w io.Writer) tx {
	ch := make(chan txpkg)
	done := make(chan struct{}, 1)

	go func() {
	loop:
		for {
			select {
			case p := <-ch:
				var writeErr error

				if nil == p.frame {
					_, writeErr = w.Write(bytesNewLine)
				} else {
					writeErr = p.frame.Write(w)
				}
				p.err <- writeErr
			case <-done:
				break loop
			}
		}
		close(ch)
	}()

	return tx{ch, done}
}

func (x tx) stop() {
	select {
	case x.done <- struct{}{}:
	default:
	}
}

// A Handle provides thead safe methods for reading from
// and writing to a connection stream.
type Handle struct {
	tx  tx
	rx  rx
}

// Bind binds a new handle to rw. The handle is available
// for reading and writing immediately.
func Bind(rw io.ReadWriter) *Handle {
	return &Handle{
		tx: newTx(rw),
		rx: newRx(rw),
	}
}

// Send sends a frame to the output stream and is thread safe.
// Send will block until the stream is available for writing.
// To send a heartbeat to the stream, set the frame argument's
// value to nil.
func (s *Handle) Send(ctx context.Context, frame *Frame) error {
	chErr := make(chan error, 1)

	s.tx.c <- txpkg{frame, chErr}

	select {
	case txErr := <-chErr:
		if nil != txErr {
			return txErr
		}
	case <-ctx.Done():
		return ctx.Err()
	}
	close(chErr)
	return nil
}

// Receive reads a frame from the input stream and is thread safe.
// Receive will block until the next frame or heartbeat becomes
// available on the input stream or an EOF is encountered. If a
// heartbeat is encountered, both the *Frame and error return
// values will be nil. If an EOF is encountered, the *Frame
// return value will be nil, and the error return value will
// be equal to io.EOF.
func (s *Handle) Receive(ctx context.Context) (*Frame, error) {
	select {
	case p := <-s.rx.c:
		return p.frame, p.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Release will release all of the handle's existing
// resources. Release will not close the underlying
// ReadWriter. Calls to Send after calling Release
// will result in a panic.
func (s *Handle) Release() {
	s.tx.stop()
	s.rx.stop()
}
