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
			"destination:/queue/test\n" +
			"\n" +
			"some test content\x00",

			Frame{
				Command: CmdSend,
				Header: Header{
					"content-length": {"17"},
					"destination": {"/queue/test"},
				},
			},

			"some test content",
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
