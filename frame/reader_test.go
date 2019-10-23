package frame

import (
	"bytes"
	"io"
	"io/ioutil"
	"strings"
	"testing"
)

const (
	terminatedTestContent = "some test content\x00 is what this is"
	shortTestContent      = "some test content"
	completeTestContent   = "some test content is what this is"
)

func readAllSized(r io.Reader, size int) ([]byte, error) {
	totalBytes := make([]byte, 0)
	buf := make([]byte, size)

	for {
		read, rdErr := r.Read(buf)

		if read > 0 {
			totalBytes = append(totalBytes, buf[:read]...)
		}

		if nil != rdErr {

			if io.EOF == rdErr {
				break
			}
			return totalBytes, rdErr
		}
	}
	return totalBytes, nil
}

func TestDelimitedReader_Read(t *testing.T) {
	r := strings.NewReader(terminatedTestContent)
	dr := delimitReader(r, byteNull)
	read, rdErr := readAllSized(dr, 1)

	if nil != rdErr && rdErr != io.EOF {
		t.Error(rdErr)
	}
	value := string(read)

	if shortTestContent != value {
		t.Errorf("bad read. expected `%s`, got `%s`", shortTestContent, value)
	}
}

func TestDelimitedReader_Read2(t *testing.T) {
	r := strings.NewReader(completeTestContent)
	dr := delimitReader(r, byteNull)
	read, rdErr := readAllSized(dr, 1)

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
	_, bdyConnWrtErr := bdyConn.WriteString(terminatedTestContent)

	if nil != bdyConnWrtErr {
		t.Fatal(bdyConnWrtErr)
	}
	frmConn := New(CmdConnect, &bdyConn)
	frmConn.Header.Append(HdrLogin, "hello")
	frmConn.Header.Append(HdrPasscode, "world")

	frmWrtErr := frmConn.Write(&frames)

	if nil != frmWrtErr {
		t.Fatal(frmWrtErr)
	}
	var bdySend bytes.Buffer
	_, bdySendWrtErr := bdySend.WriteString("I'm sending this to you")

	if nil != bdySendWrtErr {
		t.Fatal(bdySendWrtErr)
	}
	frmSend := New(CmdSend, &bdySend)
	frmWrtErr = frmSend.Write(&frames)

	if nil != frmWrtErr {
		t.Fatal(frmWrtErr)
	}

	frm1, frmRdErr := Read(&frames)

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

	if terminatedTestContent != bdyConnStr {
		t.Fatalf("wrong frame body. expected %s, got %s", terminatedTestContent, bdyConnStr)
	}

	frm2, frm2RdErr := Read(&frames)

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
