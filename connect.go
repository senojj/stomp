package stomp

import (
	"fmt"
	"github.com/dynata/stomp/proto"
	"io/ioutil"
	"net"
	"strconv"
	"strings"
)

type Session struct {
	version     string
	id          string
	server      string
	connection  net.Conn
	txHeartBeat int
	rxHeartBeat int
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

func Connect(c net.Conn, options ...func(Option)) (*Session, error) {
	host, _, splitErr := net.SplitHostPort(c.RemoteAddr().String())

	if nil != splitErr {
		return nil, splitErr
	}
	frame := proto.NewFrame(proto.CmdConnect, nil)

	frame.Header.Set(proto.HdrHost, host)
	frame.Header.Set(proto.HdrAcceptVersion, "1.0,1.1")
	frame.Header.Set(proto.HdrHeartBeat, "0,0")

	for _, option := range options {
		option(Option(frame.Header))
	}
	_, frameWrtErr := frame.WriteTo(c)

	if nil != frameWrtErr {
		return nil, frameWrtErr
	}
	frameReader := proto.NewFrameReader(c)

	respFrame, frameRdErr := frameReader.Read()
	defer respFrame.Body.Close()

	if nil != frameRdErr {
		return nil, frameRdErr
	}

	if respFrame.Command == proto.CmdError {
		contentType, ok := respFrame.Header.Get(proto.HdrContentType)
		var err error

		if ok && contentType == "text/plain" {
			body, bodyRdErr := ioutil.ReadAll(respFrame.Body)

			if nil != bodyRdErr {
				err = fmt.Errorf("unable to read frame body: %v", bodyRdErr)
			} else {
				err = fmt.Errorf("%v", body)
			}
		} else {
			err = fmt.Errorf("frame body content type is unreadable")
		}
		return nil, err
	}

	if respFrame.Command != proto.CmdConnected {
		return nil, fmt.Errorf("unexpected frame command. expected %s, got %s", proto.CmdConnected, respFrame.Command)
	}
	version, _ := respFrame.Header.Get(proto.HdrVersion)
	sessionId, _ := respFrame.Header.Get(proto.HdrSession)
	server, _ := respFrame.Header.Get(proto.HdrServer)

	heartBeat, ok := respFrame.Header.Get(proto.HdrHeartBeat)

	if !ok {
		heartBeat = "0,0"
	}
	rx, tx, hbErr := splitHeartBeat(heartBeat)

	if nil != hbErr {
		return nil, hbErr
	}
	session := Session{
		version: version,
		id: sessionId,
		server: server,
		connection: c,
		txHeartBeat: tx,
		rxHeartBeat: rx,
	}
	return &session, nil
}

func splitHeartBeat(value string) (int, int, error) {
	beats := strings.Split(value, ",")

	if len(beats) < 2 {
		return 0, 0, fmt.Errorf("malformed heart beat header: invalid length")
	}
	rx, rxErr := strconv.Atoi(beats[0])

	if nil != rxErr {
		return 0, 0, fmt.Errorf("malformed rx heart beat header value: %v", rxErr)
	}
	tx, txErr := strconv.Atoi(beats[1])

	if nil != txErr {
		return 0, 0, fmt.Errorf("malformed tx heart beat header value: %v", txErr)
	}
	return rx, tx, nil
}
