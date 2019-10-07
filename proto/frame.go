package proto

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

type ServerFrame struct {
	Command Command
	Header  Header
	Body    io.ReadCloser
}

func (sf *ServerFrame) String() string {
	return fmt.Sprintf("{Command: %s, Header: %v}", sf.Command, sf.Header)
}

type ClientFrame struct {
	Command Command
	Header  Header
	Body    io.Reader
}

func calculateContentLength(r io.Reader) int {
	if nil != r {
		if v, ok := r.(MeasurableReader); ok {
			return v.Len()
		}
		return -1
	}
	return 0
}

func (f *ClientFrame) WriteTo(writer io.Writer) (int64, error) {
	w := bufio.NewWriterSize(writer, 4096)
	var written int64 = 0

	cmdBytes, cmdWrtErr := f.Command.WriteTo(w)

	if nil != cmdWrtErr {
		return written, cmdWrtErr
	}
	written += cmdBytes

	contentLength := calculateContentLength(f.Body)

	if contentLength > -1 {
		f.Header.Set(HdrContentLength, strconv.Itoa(contentLength))
	}
	hdrBytes, hdrWrtErr := f.Header.WriteTo(w)

	if nil != hdrWrtErr {
		return written, hdrWrtErr
	}
	written += hdrBytes

	nullBytes, nullWrtErr := w.Write([]byte(charNewLine))

	if nil != nullWrtErr {
		return written, nullWrtErr
	}
	written += int64(nullBytes)

	if nil != f.Body {
		bdyb, bdyWrtErr := io.Copy(w, f.Body)

		if nil != bdyWrtErr {
			return written, bdyWrtErr
		}
		written += bdyb
	}

	nullBytes, nullWrtErr = w.Write([]byte(charNull))

	if nil != nullWrtErr {
		return written, nullWrtErr
	}
	written += int64(nullBytes)
	flErr := w.Flush()

	if nil != flErr {
		return written, flErr
	}
	return written, nil
}

func NewFrame(command Command, body io.Reader) *ClientFrame {
	return &ClientFrame{
		Command: command,
		Header:  make(Header),
		Body:    body,
	}
}
