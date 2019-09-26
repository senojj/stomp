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
	done      bool
}

func (d *DelimitedReader) Read(p []byte) (int, error) {
	if d.done {
		return 0, io.EOF
	}
	buf := make([]byte, cap(p))
	read, rdErr := d.reader.Read(buf)

	if nil != rdErr {
		return 0, rdErr
	}

	for i, v := range buf {
		p[i] = v

		if v == d.delimiter {
			d.done = true
			return i, io.EOF
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
