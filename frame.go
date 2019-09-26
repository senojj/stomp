package stomp

import (
	"io"
)

const (
	Null = "\x00"
)

type Frame struct {
	Command Command
	Header  Header
	Body    MeasurableReader
}

func (f *Frame) WriteTo(w io.Writer) (int64, error) {
	var written int64 = 0

	cmdb, cmdWrtErr := f.Command.WriteTo(w)

	if nil != cmdWrtErr {
		return written, cmdWrtErr
	}
	written += int64(cmdb)
	hdrb, hdrWrtErr := f.Header.WriteTo(w)

	if nil != hdrWrtErr {
		return written, hdrWrtErr
	}
	written += int64(hdrb)

	nlb, nlWrtErr := w.Write([]byte("\n"))

	if nil != nlWrtErr {
		return written, nlWrtErr
	}
	written += int64(nlb)

	bdyb, bdyWrtErr := io.Copy(w, f.Body)

	if nil != bdyWrtErr {
		return written, bdyWrtErr
	}
	written += int64(bdyb)

	nullb, nullWrtErr := w.Write([]byte{0})

	if nil != nullWrtErr {
		return written, nullWrtErr
	}
	written += int64(nullb)

	return written, nil
}
