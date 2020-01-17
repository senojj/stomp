package stomp

import (
	"bytes"
	"context"
	"crypto/tls"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"testing"
	"time"
)

func TestConnect(t *testing.T) {
	dialer := net.Dialer{Timeout: 5*time.Second}
	conn, dialErr := tls.DialWithDialer(&dialer, "tcp", os.Getenv("MQ_URI"), nil)

	if nil != dialErr {
		t.Fatal(dialErr)
	}
	trns := NewTransport(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	connErr := ConnectWithContext(ctx, trns, "abcd", "abcd")
	cancel()

	if nil != connErr {
		t.Fatal(connErr)
	}

	f, readErr := read(conn)

	if nil != readErr {
		t.Fatal(readErr)
	}
	var buf bytes.Buffer

	writeErr := f.Write(&buf)

	if nil != writeErr {
		t.Fatal(writeErr)
	}
	t.Log(buf.String())

	_, discardErr := buf.WriteTo(ioutil.Discard)

	if nil != discardErr {
		t.Fatal(discardErr)
	}

	f = NewFrame(CmdSend, strings.NewReader("hello there"))
	f.Header.Set(HdrDestination, "/queue/a.test")

	sendErr := f.Write(conn)

	if nil != sendErr {
		t.Fatal(sendErr)
	}

	f = NewFrame(CmdSubscribe, nil)
	f.Header.Set(HdrId, "12345")
	f.Header.Set(HdrDestination, "/queue/a.test")

	subscribeErr := f.Write(conn)

	if nil != subscribeErr {
		t.Fatal(subscribeErr)
	}

	f, readErr = read(conn)

	if nil != readErr {
		t.Fatal(readErr)
	}

	writeErr = f.Write(&buf)

	if nil != writeErr {
		t.Fatal(writeErr)
	}
	t.Log(buf.String())

	_, discardErr = buf.WriteTo(ioutil.Discard)

	if nil != discardErr {
		t.Fatal(discardErr)
	}
	conn.Close()
}

func read(r io.Reader) (*Frame, error) {
	f, readErr := ReadFrame(r)

	if nil != readErr {
		return nil, readErr
	}

	if nil == f {
		return read(r)
	}
	return f, nil
}
