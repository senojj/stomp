package proto

import (
	"fmt"
	"io"
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

func (cf *ClientFrame) String() string {
	return fmt.Sprintf("{Command: %s, Header: %v}", cf.Command, cf.Header)
}

func NewFrame(command Command, body io.Reader) *ClientFrame {
	return &ClientFrame{
		Command: command,
		Header:  make(Header),
		Body:    body,
	}
}
