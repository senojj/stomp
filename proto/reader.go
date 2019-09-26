package proto

import (
	"bytes"
	"io"
)

type MeasurableReader interface {
	io.Reader
	Len() int
}

type DelimitedReader struct {
	delimiter byte
	reader    io.Reader
	buf       bytes.Buffer
}

func (d *DelimitedReader) Read(p []byte) (int, error) {
	read, rdErr := d.reader.Read(p)

	if nil != rdErr {
		return d.buf.Len(), rdErr
	}

	for _, i := range p {
		d.buf.WriteByte(i)

		if i == d.delimiter {
			return d.buf.Len(), io.EOF
		}
	}
	return read, rdErr
}

func NewDelimitedReader(r io.Reader, delimiter byte) *DelimitedReader {
	return &DelimitedReader{
		delimiter: delimiter,
		reader:    r,
	}
}
