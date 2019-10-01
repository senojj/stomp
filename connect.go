package stomp

import (
	"context"
	"fmt"
	"github.com/dynata/stomp/proto"
	"io"
	"io/ioutil"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
)

type writeRequest struct {
	Frame *proto.ClientFrame
	C     chan<- error
}

type processor struct {
	C    chan<- writeRequest
	done chan struct{}
	wg   sync.WaitGroup
}

func (i *processor) Close() error {
	select {
	case i.done <- struct{}{}:
	default:
	}
	i.wg.Wait()
	return nil
}

func process(writer io.Writer, reader *proto.FrameReader) *processor {
	ch := make(chan writeRequest)
	done := make(chan struct{}, 1)
	receipts := make(map[string]chan<- error)
	var wg sync.WaitGroup

	go func() {
		wg.Add(1)
		defer wg.Done()

	loop:
		for {
			frame, rdErr := reader.Read()

			if nil != rdErr {
				log.Println(rdErr)
			}

			if nil != frame {
				switch frame.Command {
				case proto.CmdReceipt:
					id, ok := frame.Header.Get(proto.HdrReceiptId)

					if ok {
						receipts[id] <- nil
						delete(receipts, id)
					} else {
						log.Printf("error: stomp: unknown receipt-id: %s", id)
					}
					frame.Body.Close()
				case proto.CmdError:
					id, ok := frame.Header.Get(proto.HdrReceiptId)
					content, rdErr := ioutil.ReadAll(frame.Body)

					if nil != rdErr {
						log.Println(rdErr)
						continue
					}

					if ok {
						receipts[id] <- fmt.Errorf(string(content))
						delete(receipts, id)
					} else {
						log.Printf(string(content))
					}
					frame.Body.Close()
				}
			}

			select {
			case wr := <-ch:
				id, ok := wr.Frame.Header.Get(proto.HdrReceipt)

				if ok {
					receipts[id] = wr.C
				}
				_, wrErr := wr.Frame.WriteTo(writer)
				wr.C <- wrErr
			case <-done:
				break loop
			}
		}
	}()
	return &processor{C: ch, done: done}
}

type Session struct {
	version     string
	id          string
	server      string
	connection  net.Conn
	processor   *processor
	txHeartBeat int
	rxHeartBeat int
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

func (s *Session) Send(ctx context.Context, destination string, content io.Reader, options ...func(Option)) error {
	frame := proto.NewFrame(proto.CmdSend, content)

	for _, option := range options {
		option(Option(frame.Header))
	}
	frame.Header.Set(proto.HdrDestination, destination)
	ch := make(chan error, 1)

	req := writeRequest{frame, ch}

	select {
	case s.processor.C <- req:
		select {
		case result := <-ch:
			if nil != result {
				return result
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	case <-ctx.Done():
		return ctx.Err()
	}

	_, ok := frame.Header.Get(proto.HdrReceipt)

	if !ok {
		return nil
	}

	select {
	case result := <-ch:
		return result
	case <-ctx.Done():
		return ctx.Err()
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
	proc := process(c, frameReader)

	session := Session{
		version:     version,
		id:          sessionId,
		server:      server,
		connection:  c,
		txHeartBeat: tx,
		rxHeartBeat: rx,
		processor:   proc,
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
