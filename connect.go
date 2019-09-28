package stomp

import (
	"github.com/dynata/stomp/proto"
	"net"
)

type Connection struct {
	version string
	session string
	server  string
}

type Option proto.Header

func (o Option) Set(name string, value ...string) {
	o[name] = value
}

func WithCredentials(login, passcode string) func(Option) {
	return func(option Option) {
		option.Set(proto.HdrLogin, login)
		option.Set(proto.HdrPasscode, passcode)
	}
}

func Connect(c net.Conn, options ...func(Option)) (*Connection, error) {
	host, _, splitErr := net.SplitHostPort(c.RemoteAddr().String())

	if nil != splitErr {
		return nil, splitErr
	}
	frame := proto.NewFrame(proto.CmdConnect, nil)

	frame.Header.Set(proto.HdrHost, host)

	for _, option := range options {
		option(Option(frame.Header))
	}

	return nil, nil
}
