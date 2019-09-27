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