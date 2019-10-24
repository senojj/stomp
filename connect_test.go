package stomp

import (
	"bytes"
	"crypto/tls"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestConnect(t *testing.T) {
	conn, dialErr := tls.Dial("tcp", os.Getenv("MQ_URI"), nil)

	if nil != dialErr {
		t.Fatal(dialErr)
	}

	f := NewFrame(CmdConnect, nil)
	f.Header.Set(HdrLogin, "mixr")
	f.Header.Set(HdrPasscode, os.Getenv("MQ_PASSWORD"))

	connErr := f.Write(conn)

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
