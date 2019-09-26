package stomp

import (
	"bytes"
	"io/ioutil"
	"testing"
)

const someTestContent = "some test content"
const frameData = "SEND\ncontent-length:17\n\nsome test content\x00"

func TestFrame_WriteTo(t *testing.T) {
	in := bytes.NewBufferString(someTestContent)

	frame := Frame{
		Command: CmdSend,
		Header: make(Header),
		Body: in,
	}
	var out bytes.Buffer
	_, wrtErr := frame.WriteTo(&out)

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
