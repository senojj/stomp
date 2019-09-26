package proto

import (
	"io"
	"strings"
	"testing"
)

const (
	terminatedTestContent = "some test content\x00 is what this is"
)

func TestDelimitedReader_Read(t *testing.T) {
	r := strings.NewReader(terminatedTestContent)
	dr := NewDelimitedReader(r, "\000")
	p := make([]byte, 1024)
	_, rdErr := dr.Read(p)

	if nil != rdErr && rdErr != io.EOF {
		t.Error(rdErr)
	}

	if "some test content\x00" != string(p) {
		t.Errorf("bad read. expected `%s`, got `%s`", "some test content\\x00", string(p))
	}
}