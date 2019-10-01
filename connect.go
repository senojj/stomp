package stomp

import (
	"fmt"
	"github.com/dynata/stomp/proto"
	"io"
	"io/ioutil"
	"net"
	"strconv"
	"strings"
	"sync"
)

type writeResult struct {
	Written int64
	Err     error
}

type writeRequest struct {
	Frame proto.ClientFrame
	C     chan<- writeResult
}

type readResult struct {
	Frame *proto.ServerFrame
	Err   error
}

type consumer struct {
	C    <-chan readResult
	done chan struct{}
	wg   sync.WaitGroup
}

func (i *consumer) Close() error {
	select {
	case i.done <- struct{}{}:
	default:
	}
	i.wg.Wait()
	return nil
}

type producer struct {
	C    chan<- writeRequest
	done chan struct{}
	wg   sync.WaitGroup
}

func (i *producer) Close() error {
	select {
	case i.done <- struct{}{}:
	default:
	}
	i.wg.Wait()
	return nil
}

func consume(r *proto.FrameReader) *consumer {
	ch := make(chan readResult)
	done := make(chan struct{}, 1)
	var wg sync.WaitGroup

	go func() {
		wg.Add(1)
		defer wg.Done()

	loop:
		for {
			frame, err := r.Read()

			select {
			case ch <- readResult{frame, err}:
			case <-done:
				break loop
			}
		}
	}()
	return &consumer{C: ch, done: done, wg: wg}
}

func produce(w io.Writer) *producer {
	ch := make(chan writeRequest)
	done := make(chan struct{}, 1)
	var wg sync.WaitGroup

	go func() {
		wg.Add(1)
		defer wg.Done()

	loop:
		for {
			select {
			case req, ok := <-ch:
				if !ok {
					break loop
				}
				written, err := req.Frame.WriteTo(w)
				req.C <- writeResult{written, err}
			case <-done:
				break loop
			}
		}
	}()
	return &producer{C: ch, done: done, wg: wg}
}

type Session struct {
	version        string
	id             string
	server         string
	connection     net.Conn
	consumer       *consumer
	producer       *producer
	txHeartBeat    int
	rxHeartBeat    int
}

func (s *Session) String() string {
	return fmt.Sprintf(
		"{Version: %s, ID: %s, Server: %s, TxHeartBeat: %d, RxHeartBeat: %d}",
		s.version,
		s.id,
		s.server,
		s.txHeartBeat,
		s.rxHeartBeat,
	)
}

func (s *Session) Send(destination string, options ...Option) error {
	return nil
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

func WithHeartBeat(tx, rx int) func(Option) {
	return func(option Option) {
		option.Set(proto.HdrHeartBeat, strconv.Itoa(tx)+","+strconv.Itoa(rx))
	}
}

func Connect(c net.Conn, options ...func(Option)) (*Session, error) {
	host, _, splitErr := net.SplitHostPort(c.RemoteAddr().String())

	if nil != splitErr {
		return nil, splitErr
	}
	frame := proto.NewFrame(proto.CmdConnect, nil)

	frame.Header.Set(proto.HdrHost, host)
	frame.Header.Set(proto.HdrAcceptVersion, "1.1,1.2")
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

	if nil != frameRdErr {
		return nil, frameRdErr
	}
	defer respFrame.Body.Close()

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
	tx, rx, hbErr := splitHeartBeat(heartBeat)

	if nil != hbErr {
		return nil, hbErr
	}
	cons := consume(frameReader)
	prod := produce(c)

	session := Session{
		version:     version,
		id:          sessionId,
		server:      server,
		connection:  c,
		txHeartBeat: tx,
		rxHeartBeat: rx,
		consumer: cons,
		producer: prod,
	}
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
