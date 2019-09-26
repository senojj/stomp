package frame

import (
	"github.com/dynata/stomp/frame/command"
	"github.com/dynata/stomp/frame/header"
	"io"
)

type Frame struct {
	Command command.Command
	Headers header.Map
	Body io.Reader
}

func (f *Frame) WriteTo(w io.Writer) (int64, error) {
	var written int64 = 0

	cmdb, cmdWrtErr := f.Command.WriteTo(w)

	if nil != cmdWrtErr {
		return written, cmdWrtErr
	}
	written += int64(cmdb)
	hdrb, hdrWrtErr := f.Headers.WriteTo(w)

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

	return written, nil
}
