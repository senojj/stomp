package stomp

import (
	"context"
	"fmt"
	"github.com/dynata/stomp/proto"
	"io"
	"net"
	"sync"
	"time"
)

type Session struct {
	Version     string
	ID          string
	Server      string
	connection  net.Conn
	processor   *processor
	txHeartBeat int
	rxHeartBeat int
	iden        *uint64
	m           sync.Mutex
	closed      bool
}

func (s *Session) String() string {
	return fmt.Sprintf(
		"{Version: %s, ID: %s, Server: %s, TxHeartBeat: %d, RxHeartBeat: %d}",
		s.Version,
		s.ID,
		s.Server,
		s.txHeartBeat,
		s.rxHeartBeat,
	)
}

func (s *Session) Close() error {
	s.m.Lock()
	defer s.m.Unlock()

	if s.closed {
		return nil
	}
	frame := proto.NewFrame(proto.CmdDisconnect, nil)
	frame.Header.Set(proto.HdrReceipt, "session-disconnect")
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	sndErr := s.sendFrame(ctx, frame)
	cancel()
	s.processor.Close()
	s.closed = true
	return sndErr
}

func (s *Session) sendFrame(ctx context.Context, frame *proto.ClientFrame, args ...interface{}) error {
	ch := make(chan error, 1)

	req := writeRequest{frame, ch, args}

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

func (s *Session) Send(
	ctx context.Context,
	destination string,
	content io.Reader,
	options ...func(Option),
) error {
	s.m.Lock()
	defer s.m.Unlock()

	if s.closed {
		return ErrSessionClosed
	}

	frame := proto.NewFrame(proto.CmdSend, content)

	for _, option := range options {
		option(Option(frame.Header))
	}
	frame.Header.Set(proto.HdrDestination, destination)
	return s.sendFrame(ctx, frame)
}

func (s *Session) Subscribe(
	ctx context.Context,
	destination string,
	fn func(Message),
	options ...func(Option),
) (*Subscription, error) {
	if s.closed {
		return nil, ErrSessionClosed
	}
	frame := proto.NewFrame(proto.CmdSubscribe, nil)

	for _, option := range options {
		option(Option(frame.Header))
	}
	id := nextId()
	frame.Header.Set(proto.HdrId, id)
	frame.Header.Set(proto.HdrDestination, destination)

	sendErr := s.sendFrame(ctx, frame, fn)

	if nil != sendErr {
		return nil, sendErr
	}

	return &Subscription{
		id:      id,
		session: s,
	}, nil
}

type Subscription struct {
	id      string
	session *Session
}

func (s *Subscription) Unsubscribe(ctx context.Context, options ...func(Option)) error {
	if s.session.closed {
		return ErrSessionClosed
	}
	frame := proto.NewFrame(proto.CmdUnsubscribe, nil)

	for _, option := range options {
		option(Option(frame.Header))
	}
	frame.Header.Set(proto.HdrId, s.id)
	return s.session.sendFrame(ctx, frame)
}
