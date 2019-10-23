package stomp

import (
	"fmt"
	"github.com/dynata/stomp/frame"
	"io/ioutil"
	"net"
	"strconv"
	"strings"
)

func Connect(c net.Conn, options ...func(Option)) (*Session, error) {
	host, _, splitErr := net.SplitHostPort(c.RemoteAddr().String())

	if nil != splitErr {
		return nil, splitErr
	}
	f := frame.New(frame.CmdConnect, nil)

	f.Header.Set(frame.HdrHost, host)
	f.Header.Set(frame.HdrAcceptVersion, "1.1,1.2")
	f.Header.Set(frame.HdrHeartBeat, "0,0")

	for _, option := range options {
		option(Option(f.Header))
	}
	frameWrtErr := f.Write(c)

	if nil != frameWrtErr {
		return nil, frameWrtErr
	}

	respFrame, frameRdErr := frame.Read(c)

	if nil != frameRdErr {
		return nil, frameRdErr
	}

	if respFrame.Command == frame.CmdError {
		contentType, ok := respFrame.Header.Get(frame.HdrContentType)
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

	if respFrame.Command != frame.CmdConnected {
		return nil, fmt.Errorf("unexpected frame command. expected %s, got %s", frame.CmdConnected, respFrame.Command)
	}
	version, _ := respFrame.Header.Get(frame.HdrVersion)
	sessionId, _ := respFrame.Header.Get(frame.HdrSession)
	server, _ := respFrame.Header.Get(frame.HdrServer)

	heartBeat, ok := respFrame.Header.Get(frame.HdrHeartBeat)

	if !ok {
		heartBeat = "0,0"
	}
	tx, rx, hbErr := splitHeartBeat(heartBeat)

	if nil != hbErr {
		return nil, hbErr
	}
	respFrame.Body.Close()
	proc := process(c)

	session := Session{
		Version:     version,
		ID:          sessionId,
		Server:      server,
		txHeartBeat: tx,
		rxHeartBeat: rx,
		processor:   proc,
	}

	go func() {
		for p := range proc.R {
			ch, has := session.subscriptions.Get(p.ID)

			if has {
				ch <- p.Message
			} else {
				p.Message.Body.Close()
			}
		}
	}()
	return &session, nil
}

func splitHeartBeat(value string) (int, int, error) {
	beats := strings.Split(value, ",")

	if len(beats) < 2 {
		return 0, 0, fmt.Errorf("malformed heart beat header: invalid length")
	}
	tx, txErr := strconv.Atoi(beats[0])

	if nil != txErr {
		return 0, 0, fmt.Errorf("malformed rx heart beat header value: %v", txErr)
	}
	rx, rxErr := strconv.Atoi(beats[1])

	if nil != rxErr {
		return 0, 0, fmt.Errorf("malformed tx heart beat header value: %v", rxErr)
	}
	return tx, rx, nil
}
