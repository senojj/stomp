package proto

import (
	"io"
	"strconv"
)

type Frame struct {
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

func (f *Frame) WriteTo(w io.Writer) (int64, error) {
	var written int64 = 0

	cmdb, cmdWrtErr := f.Command.WriteTo(w)

	if nil != cmdWrtErr {
		return written, cmdWrtErr
	}
	written += cmdb

	contentLength := calculateContentLength(f.Body)

	if contentLength > -1 {
		f.Header.Set(HdrContentLength, strconv.Itoa(contentLength))
	}
	hdrb, hdrWrtErr := f.Header.WriteTo(w)

	if nil != hdrWrtErr {
		return written, hdrWrtErr
	}
	written += hdrb

	nlb, nlWrtErr := w.Write([]byte(charNewLine))

	if nil != nlWrtErr {
		return written, nlWrtErr
	}
	written += int64(nlb)

	bdyb, bdyWrtErr := io.Copy(w, f.Body)

	if nil != bdyWrtErr {
		return written, bdyWrtErr
	}
	written += bdyb

	nullb, nullWrtErr := w.Write([]byte(charNull))

	if nil != nullWrtErr {
		return written, nullWrtErr
	}
	written += int64(nullb)

	return written, nil
}

func NewFrame(command Command, body io.Reader) *Frame {
	return &Frame{
		Command: command,
		Header: make(Header),
		Body: body,
	}
}
