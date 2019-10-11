package stomp

import (
	"context"
	"fmt"
	"github.com/dynata/stomp/proto"
	"io"
	"sync"
	"time"
)

type Session struct {
	Version       string
	ID            string
	Server        string
	processor     *processor
	txHeartBeat   int
	rxHeartBeat   int
	m             sync.Mutex
	closed        bool
	subscriptions subscriptionMap
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

	req := request{frame, ch, args}

	select {
	case s.processor.W <- req:
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

func (s *Session) Ack(ctx context.Context, msg Message, options ...func(Option)) error {
	s.m.Lock()
	defer s.m.Unlock()

	if s.closed {
		return ErrSessionClosed
	}

	frame := proto.NewFrame(proto.CmdAck, nil)

	for _, option := range options {
		option(Option(frame.Header))
	}

	switch s.Version {
	case v12:
		ack, ok := msg.Header.Get(proto.HdrAck)

		if ok {
			frame.Header.Set(proto.HdrId, ack)
		}
	case v10, v11:
		subId, ok := msg.Header.Get(proto.HdrSubscription)

		if ok {
			frame.Header.Set(proto.HdrSubscription, subId)
		}
		msgId, ok := msg.Header.Get(proto.HdrMessageId)

		if ok {
			frame.Header.Set(proto.HdrMessageId, msgId)
		}
	}
	return s.sendFrame(ctx, frame)
}

func (s *Session) Nack(ctx context.Context, msg Message, options ...func(Option)) error {
	s.m.Lock()
	defer s.m.Unlock()

	if s.closed {
		return ErrSessionClosed
	}

	frame := proto.NewFrame(proto.CmdNack, nil)

	for _, option := range options {
		option(Option(frame.Header))
	}

	switch s.Version {
	case v12:
		ack, ok := msg.Header.Get(proto.HdrAck)

		if ok {
			frame.Header.Set(proto.HdrId, ack)
		}
	case v10, v11:
		subId, ok := msg.Header.Get(proto.HdrSubscription)

		if ok {
			frame.Header.Set(proto.HdrSubscription, subId)
		}
		msgId, ok := msg.Header.Get(proto.HdrMessageId)

		if ok {
			frame.Header.Set(proto.HdrMessageId, msgId)
		}
	}
	return s.sendFrame(ctx, frame)
}

func (s *Session) Begin(ctx context.Context, options ...func(Option)) (*Transaction, error) {
	frame := proto.NewFrame(proto.CmdBegin, nil)

	beginErr := s.sendFrame(ctx, frame)

	for _, option := range options {
		option(Option(frame.Header))
	}

	trnId, ok := frame.Header.Get(proto.HdrTransaction)

	if !ok {
		trnId = nextId()
		frame.Header.Set(proto.HdrTransaction, trnId)
	}

	if nil != beginErr {
		return nil, beginErr
	}
	return &Transaction{trnId, s}, nil
}

func (s *Session) Subscribe(
	ctx context.Context,
	destination string,
	options ...func(Option),
) (*Subscription, error) {
	if s.closed {
		return nil, ErrSessionClosed
	}
	frame := proto.NewFrame(proto.CmdSubscribe, nil)
	frame.Header.Set(proto.HdrAck, AckAuto)

	for _, option := range options {
		option(Option(frame.Header))
	}
	id := nextId()
	frame.Header.Set(proto.HdrId, id)
	frame.Header.Set(proto.HdrDestination, destination)

	sendErr := s.sendFrame(ctx, frame)

	if nil != sendErr {
		return nil, sendErr
	}
	subCh := make(chan Message)
	s.subscriptions.Set(id, subCh)

	return &Subscription{
		id:      id,
		session: s,
		C:       subCh,
	}, nil
}

type Subscription struct {
	id      string
	session *Session
	C       <-chan Message
}

func (s *Subscription) Unsubscribe(ctx context.Context) error {
	if s.session.closed {
		return ErrSessionClosed
	}
	frame := proto.NewFrame(proto.CmdUnsubscribe, nil)

	frame.Header.Set(proto.HdrId, s.id)
	frame.Header.Set(proto.HdrReceipt, nextId())
	sndErr := s.session.sendFrame(ctx, frame)

	if nil != sndErr {
		return sndErr
	}
	ch, has := s.session.subscriptions.Get(s.id)

	if has {
		s.session.subscriptions.Del(s.id)
		close(ch)
	}
	return nil
}
