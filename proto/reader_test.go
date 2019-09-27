package proto

import (
	"bytes"
	"io"
	"io/ioutil"
	"strings"
	"testing"
)

const (
	terminatedTestContent = "some test content\x00 is what this is"
)

func TestDelimitedReader_Read(t *testing.T) {
	r := strings.NewReader(terminatedTestContent)
	dr := DelimitReader(r, byteNull)
	p := make([]byte, 1024)
	read, rdErr := dr.Read(p)

	if nil != rdErr && rdErr != io.EOF {
		t.Error(rdErr)
	}
	value := string(p[:read])

	if "some test content" != value {
		t.Errorf("bad read. expected `%s`, got `%s`", "some test content", value)
	}
}

func TestFrameReader_Read(t *testing.T) {
	var frames bytes.Buffer
	var bdyConn bytes.Buffer
	_, bdyConnWrtErr := bdyConn.WriteString("hello world!")

	if nil != bdyConnWrtErr {
		t.Fatal(bdyConnWrtErr)
	}
	frmConn := NewFrame(CmdConnect, &bdyConn)
	frmConn.Header.Append(HdrLogin, "hello")
	frmConn.Header.Append(HdrPasscode, "world")

	_, frmWrtErr := frmConn.WriteTo(&frames)

	if nil != frmWrtErr {
		t.Fatal(frmWrtErr)
	}
	var bdySend bytes.Buffer
	_, bdySendWrtErr := bdySend.WriteString("I'm sending this to you")

	if nil != bdySendWrtErr {
		t.Fatal(bdySendWrtErr)
	}
	frmSend := NewFrame(CmdSend, &bdySend)

	_, frmWrtErr = frmSend.WriteTo(&frames)

	if nil != frmWrtErr {
		t.Fatal(frmWrtErr)
	}
	frmReader := NewFrameReader(&frames)

	frm1, frmRdErr := frmReader.Read()

	if nil != frmRdErr {
		t.Fatal(frmRdErr)
	}

	if frm1.Command != CmdConnect {
		t.Fatalf("wrong frame command. expected %s, got %s.", CmdConnect, frm1.Command)
	}

	bdyConnBytes, bdyConnRdErr := ioutil.ReadAll(frm1.Body)

	if nil != bdyConnRdErr {
		t.Fatal(bdyConnRdErr)
	}

	bdyConnStr := string(bdyConnBytes)

	if "hello world!" != bdyConnStr {
		t.Fatalf("wrong frame body. expected %s, got %s.", "hello world!", bdyConnStr)
	}
}
