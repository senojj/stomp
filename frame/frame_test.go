package frame

import (
	"bytes"
	"fmt"
	"go/token"
	"io"
	"reflect"
	"strings"
	"testing"
)

type frameTest struct {
	Raw   string
	Frame Frame
	Body  string
}

var frameTests = []frameTest{
	{
		"SEND\n" +
			"content-length:17\n" +
			"content-type:text/plain\n" +
			"destination:/queue/test\n" +
			"\n" +
			"some test content\x00",

		Frame{
			Command: CmdSend,
			Header: Header{
				HdrContentLength: {"17"},
				HdrContentType:   {"text/plain"},
				HdrDestination:   {"/queue/test"},
			},
		},

		"some test content",
	},
	{
		"MESSAGE\n" +
			"subscription:0\n" +
			"message-id:007\n" +
			"destination:/queue/test\n" +
			"content-type:text/plain\n" +
			"\n" +
			"hello queue test\x00",

		Frame{
			Command: CmdMessage,
			Header: Header{
				HdrSubscription: {"0"},
				HdrMessageId:    {"007"},
				HdrDestination:  {"/queue/test"},
				HdrContentType:  {"text/plain"},
			},
		},

		"hello queue test",
	},
	{
		"RECEIPT\n" +
			"receipt-id:message-12345\n" +
			"\n" +
			"\x00",

		Frame{
			Command: CmdReceipt,
			Header: Header{
				HdrReceiptId: {"message-12345"},
			},
		},

		"",
	},
	{
		"ERROR\n" +
			"receipt-id:message-12345\n" +
			"content-type:text/plain\n" +
			"content-length:171\n" +
			"message: malformed frame received\n" +
			"\n" +
			"The message:\n" +
			"-----\n" +
			"MESSAGE\n" +
			"destined:/queue/a\n" +
			"receipt:message-12345\n" +
			"\n" +
			"Hello queue a!\n" +
			"-----\n" +
			"Did not contain a destination header, which is REQUIRED\n" +
			"for message propagation.\n" +
			"\x00",

		Frame{
			Command: CmdError,
			Header: Header{
				HdrReceiptId:     {"message-12345"},
				HdrContentType:   {"text/plain"},
				HdrContentLength: {"171"},
				HdrMessage:       {" malformed frame received"},
			},
		},

		"The message:\n" +
			"-----\n" +
			"MESSAGE\n" +
			"destined:/queue/a\n" +
			"receipt:message-12345\n" +
			"\n" +
			"Hello queue a!\n" +
			"-----\n" +
			"Did not contain a destination header, which is REQUIRED\n" +
			"for message propagation.\n",
	},
}

func TestReadFrame(t *testing.T) {
	for i, tt := range frameTests {
		f, readErr := Read(strings.NewReader(tt.Raw))

		if nil != readErr {
			t.Errorf("#%d: %v", i, readErr)
		}
		fbody := f.Body
		f.Body = nil
		diff(t, fmt.Sprintf("#%d Frame", i), f, &tt.Frame)
		var bout bytes.Buffer

		if nil != fbody {
			_, copyErr := io.Copy(&bout, fbody)

			if nil != copyErr {
				t.Errorf("#%d: %v", i, copyErr)
				continue
			}
			fbody.Close()
		}
		body := bout.String()

		if body != tt.Body {
			t.Errorf("#%d: Body = %q want %q", i, body, tt.Body)
		}
	}
}

func TestWriteFrame(t *testing.T) {
	for i, tt := range frameTests {
		f, readErr := Read(strings.NewReader(tt.Raw))

		if nil != readErr {
			t.Errorf("#%d: %v", i, readErr)
			continue
		}
		var buf bytes.Buffer
		writeErr := f.Write(&buf)

		if nil != writeErr {
			t.Errorf("#%d: %v", i, writeErr)
			continue
		}
	}
}

func diff(t *testing.T, prefix string, have, want interface{}) {
	hv := reflect.ValueOf(have).Elem()
	wv := reflect.ValueOf(want).Elem()
	if hv.Type() != wv.Type() {
		t.Errorf("%s: type mismatch %v want %v", prefix, hv.Type(), wv.Type())
	}
	for i := 0; i < hv.NumField(); i++ {
		name := hv.Type().Field(i).Name
		if !token.IsExported(name) {
			continue
		}
		hf := hv.Field(i).Interface()
		wf := wv.Field(i).Interface()
		if !reflect.DeepEqual(hf, wf) {
			t.Errorf("%s: %s = %v want %v", prefix, name, hf, wf)
		}
	}
}
