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
	completeTestContent = "some test content is what this is"
)

func TestDelimitedReader_Read(t *testing.T) {
	r := strings.NewReader(terminatedTestContent)
	dr := DelimitReader(r, byteNull)
	read, rdErr := ioutil.ReadAll(dr)

	if nil != rdErr && rdErr != io.EOF {
		t.Error(rdErr)
	}
	value := string(read)

	if "some test content" != value {
		t.Errorf("bad read. expected `%s`, got `%s`", "some test content", value)
	}
}

func TestDelimitedReader_Read2(t *testing.T) {
	r := strings.NewReader(completeTestContent)
	dr := DelimitReader(r, byteNull)
	read, rdErr := ioutil.ReadAll(dr)

	if nil != rdErr && rdErr != io.EOF {
		t.Error(rdErr)
	}
	value := string(read)

	if completeTestContent != value {
		t.Errorf("bad read. expected `%s`, got `%s`", completeTestContent, value)
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
		t.Fatalf("wrong frame command. expected %s, got %s", CmdConnect, frm1.Command)
	}

	bdyConnBytes, bdyConnRdErr := ioutil.ReadAll(frm1.Body)

	if nil != bdyConnRdErr {
		t.Fatal(bdyConnRdErr)
	}

	bdyConnStr := string(bdyConnBytes)

	if "hello world!" != bdyConnStr {
		t.Fatalf("wrong frame body. expected %s, got %s", "hello world!", bdyConnStr)
	}

	frm2, frm2RdErr := frmReader.Read()

	if nil != frm2RdErr {
		t.Fatal(frm2RdErr)
	}

	if frm2.Command != CmdSend {
		t.Fatalf("wrong frame command. expected %v, got %v", []byte(CmdSend), []byte(frm2.Command))
	}

	bdySendBytes, bdySendRdErr := ioutil.ReadAll(frm2.Body)

	if nil != bdySendRdErr {
		t.Fatal(bdySendRdErr)
	}

	bdySendStr := string(bdySendBytes)

	if "I'm sending this to you" != bdySendStr {
		t.Fatalf("wrong frame body. expected %s, got %s", "I'm sending this to you", bdySendStr)
	}
}
