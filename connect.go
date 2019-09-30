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
	frame.Header.Set(proto.HdrAcceptVersion, "1.0,1.1")

	for _, option := range options {
		option(Option(frame.Header))
	}
	_, frameWrtErr := frame.WriteTo(c)

	if nil != frameWrtErr {
		return nil, frameWrtErr
	}
	frameReader := proto.NewFrameReader(c)

	respFrame, frameRdErr := frameReader.Read()

	if nil != frameRdErr {
		return nil, frameRdErr
	}

	if respFrame.Command == proto.CmdError {
		// TODO: handle error
		return nil, nil
	}

	if respFrame.Command != proto.CmdConnected {
		// TODO: return error
		return nil, nil
	}
	// TODO: handle Connected frame
	return nil, nil
}
