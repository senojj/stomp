package proto

import (
	"bufio"
	"bytes"
	"fmt"
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

type FrameReader struct {
	reader *bufio.Reader
}

func (fr *FrameReader) Read() (*Frame, error) {
	cmdLine, cmdLineRdErr := fr.reader.ReadBytes(byteNewLine)

	if nil != cmdLineRdErr {
		return nil, cmdLineRdErr
	}

	if len(cmdLine) == 0 {
		return nil, fmt.Errorf("empty command")
	}
	command := Command(cmdLine)
	header := make(Header)

	for {
		hdrLine, hdrLineErr := fr.reader.ReadBytes(byteNewLine)

		if nil != hdrLineErr {
			return nil, cmdLineRdErr
		}

		if len(hdrLine) == 0 {
			break
		}
		ndx := bytes.IndexByte(hdrLine, byteColon)

		if ndx <= 0 {
			return nil, fmt.Errorf("malformed header")
		}
		name := decode(string(hdrLine[0:ndx]))
		value := decode(string(hdrLine[:ndx+1]))
		header.Append(name, value)
	}
	// TODO: check content length
}

func NewFrameReader(r io.Reader) *FrameReader {
	return &FrameReader{
		reader: bufio.NewReader(r),
	}
}
