package proto

import (
	"io"
	"strconv"
)

type ServerFrame struct {
	Command Command
	Header  Header
	Body    io.ReadCloser
}

type ClientFrame struct {
	Command Command
	Header  Header
	Body    io.Reader
}

func calculateContentLength(r io.Reader) int {
	if nil != r {
		v, ok := r.(MeasurableReader)

		if ok {
			return v.Len()
		}
		return -1
	} else {
		return 0
	}
}

func (f *ClientFrame) WriteTo(w io.Writer) (int64, error) {
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

	return written, nil
}

func NewFrame(command Command, body io.Reader) *ClientFrame {
	return &ClientFrame{
		Command: command,
		Header:  make(Header),
		Body:    body,
	}
}
