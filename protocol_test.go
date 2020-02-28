package stomp

import (
	"context"
	"errors"
	"io"
	"net"
	"strings"
	"testing"
	"time"
)

var errNoConnection = errors.New("connection not yet established")
var errBadVersion = errors.New("unsupported version")
var errInvalidLogin = errors.New("invalid login")
var errInvalidPasscode = errors.New("invalid passcode")

func listenConnect(t *testing.T, rw io.ReadWriter) {

	go func() {
	loop:
		for {
			frm, readFrameErr := ReadFrame(rw)

			if nil != readFrameErr {

				if readFrameErr == io.EOF {
					break loop
				}
				t.Fatal(readFrameErr)
			}
			closeErr := frm.Body.Close()

			if nil != closeErr {
				t.Fatal(closeErr)
			}

			var err error

			if frm.Command != CmdConnect {
				err = errNoConnection
			} else {
				version := "1.0"

				if v, ok := frm.Header.Get(HdrAcceptVersion); ok {
					versions := strings.Split(v, ",")

					for _, version = range versions {

						if version == "1.1" {
							break
						}
					}
				}

				if version != "1.1" {
					err = errBadVersion
				} else {

					if _, ok := frm.Header.Get(HdrLogin); !ok {
						err = errInvalidLogin
					} else {

						if _, ok := frm.Header.Get(HdrPasscode); !ok {
							err = errInvalidPasscode
						}
					}
				}
			}

			var frmOut *Frame

			if nil != err {
				frmOut = NewFrame(CmdError, nil)
				frmOut.Header.Set(HdrMessage, err.Error())

				if receipt, ok := frm.Header.Get(HdrReceipt); ok {
					frmOut.Header.Set(HdrReceiptId, receipt)
				}
			} else {
				frmOut = NewFrame(CmdConnected, nil)
			}
			frmOut.Header.Set(HdrVersion, "1.1")
			writeErr := frmOut.Write(rw)

			if nil != writeErr {
				t.Fatal(writeErr)
			}
		}
	}()
}

func TestConnect(t *testing.T) {
	srv, client := net.Pipe()
	listenConnect(t, srv)

	handle := Bind(client)

	f := NewFrame(CmdConnect, nil)
	f.Header.Set(HdrLogin, "test-user")
	f.Header.Set(HdrPasscode, "test-password")
	f.Header.Set(HdrAcceptVersion, "1.1")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	sendErr := handle.Send(ctx, f)

	if nil != sendErr {
		t.Fatal(sendErr)
	}
	resp, readErr := handle.Receive(ctx)
	cancel()

	if nil != readErr {
		t.Fatal(readErr)
	}

	if nil == resp {
		t.Fatal("empty frame")
	}
	closeErr := resp.Body.Close()

	if nil != closeErr {
		t.Fatal(closeErr)
	}

	if resp.Command != CmdConnected {
		message, ok := resp.Header["message"]

		if ok {
			t.Fatal(errors.New(strings.Join(message, ":")))
		}
		t.Fatal(errors.New("unknown error"))
	}
}
