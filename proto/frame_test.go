package proto

import (
	"bytes"
	"io/ioutil"
	"testing"
)

const someTestContent = "some test content"
const frameData = "SEND\ncontent-length:17\n\nsome test content\x00"

func TestFrame_WriteTo(t *testing.T) {
	in := bytes.NewBufferString(someTestContent)

	frame := NewFrame(CmdSend, in)
	var out bytes.Buffer
	frameWriter := NewFrameWriter(&out)
	_, wrtErr := frameWriter.Write(frame)

	if nil != wrtErr {
		t.Error(wrtErr)
	}
	data, rdErr := ioutil.ReadAll(&out)

	if nil != rdErr {
		t.Error(rdErr)
	}
	if string(data) != frameData {
		t.Errorf("data does not match.\nexpected:\n%s\ngot:\n%s", frameData, string(data))
	}
}
