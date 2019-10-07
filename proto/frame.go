package proto

import (
	"fmt"
	"io"
)

type Frame interface {
	Command() Command
	Header() Header
	Body() io.Reader
	io.Closer
}

type serverFrame struct {
	command Command
	header  Header
	body    io.ReadCloser
}

func (sf *serverFrame) Command() Command {
	return sf.command
}

func (sf *serverFrame) Header() Header {
	return sf.header
}

func (sf *serverFrame) Body() io.Reader {
	return sf.body
}

func (sf *serverFrame) String() string {
	return fmt.Sprintf("{Command: %s, Header: %v}", sf.Command(), sf.Header())
}

func (sf *serverFrame) Close() error {
	return sf.body.Close()
}

type clientFrame struct {
	command Command
	header  Header
	body    io.Reader
}

func (cf *clientFrame) Command() Command {
	return cf.command
}

func (cf *clientFrame) Header() Header {
	return cf.header
}

func (cf *clientFrame) Body() io.Reader {
	return cf.body
}

func (cf *clientFrame) Close() error {
	return nil
}

func NewFrame(command Command, body io.Reader) *clientFrame {
	return &clientFrame{
		command: command,
		header:  make(Header),
		body:    body,
	}
}
