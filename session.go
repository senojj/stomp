package stomp

import (
	"context"
	"fmt"
	"github.com/dynata/stomp/frame"
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
	f := frame.New(frame.CmdDisconnect, nil)
	f.Header.Set(frame.HdrReceipt, "session-disconnect")
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	sndErr := s.sendFrame(ctx, f)
	cancel()
	s.processor.Close()
	s.closed = true
	return sndErr
}

func (s *Session) sendFrame(ctx context.Context, f *frame.Frame, args ...interface{}) error {
	ch := make(chan error, 1)

	req := request{f, ch, args}

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

	_, ok := f.Header.Get(frame.HdrReceipt)

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

	f := frame.New(frame.CmdSend, content)

	for _, option := range options {
		option(Option(f.Header))
	}
	f.Header.Set(frame.HdrDestination, destination)
	return s.sendFrame(ctx, f)
}

func (s *Session) Ack(ctx context.Context, msg Message, options ...func(Option)) error {
	s.m.Lock()
	defer s.m.Unlock()

	if s.closed {
		return ErrSessionClosed
	}

	f := frame.New(frame.CmdAck, nil)

	for _, option := range options {
		option(Option(f.Header))
	}

	switch s.Version {
	case v12:
		ack, ok := msg.Header.Get(frame.HdrAck)

		if ok {
			f.Header.Set(frame.HdrId, ack)
		}
	case v10, v11:
		subId, ok := msg.Header.Get(frame.HdrSubscription)

		if ok {
			f.Header.Set(frame.HdrSubscription, subId)
		}
		msgId, ok := msg.Header.Get(frame.HdrMessageId)

		if ok {
			f.Header.Set(frame.HdrMessageId, msgId)
		}
	}
	return s.sendFrame(ctx, f)
}

func (s *Session) Nack(ctx context.Context, msg Message, options ...func(Option)) error {
	s.m.Lock()
	defer s.m.Unlock()

	if s.closed {
		return ErrSessionClosed
	}

	f := frame.New(frame.CmdNack, nil)

	for _, option := range options {
		option(Option(f.Header))
	}

	switch s.Version {
	case v12:
		ack, ok := msg.Header.Get(frame.HdrAck)

		if ok {
			f.Header.Set(frame.HdrId, ack)
		}
	case v10, v11:
		subId, ok := msg.Header.Get(frame.HdrSubscription)

		if ok {
			f.Header.Set(frame.HdrSubscription, subId)
		}
		msgId, ok := msg.Header.Get(frame.HdrMessageId)

		if ok {
			f.Header.Set(frame.HdrMessageId, msgId)
		}
	}
	return s.sendFrame(ctx, f)
}

func (s *Session) Begin(ctx context.Context, options ...func(Option)) (*Transaction, error) {
	f := frame.New(frame.CmdBegin, nil)

	for _, option := range options {
		option(Option(f.Header))
	}

	trnId, ok := f.Header.Get(frame.HdrTransaction)

	if !ok {
		trnId = nextId()
		f.Header.Set(frame.HdrTransaction, trnId)
	}

	beginErr := s.sendFrame(ctx, f)

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
	f := frame.New(frame.CmdSubscribe, nil)
	f.Header.Set(frame.HdrAck, AckAuto)

	for _, option := range options {
		option(Option(f.Header))
	}
	id := nextId()
	f.Header.Set(frame.HdrId, id)
	f.Header.Set(frame.HdrDestination, destination)

	sendErr := s.sendFrame(ctx, f)

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

func (s *Subscription) Unsubscribe(ctx context.Context, options ...func(Option)) error {
	if s.session.closed {
		return ErrSessionClosed
	}
	f := frame.New(frame.CmdUnsubscribe, nil)

	for _, option := range options {
		option(Option(f.Header))
	}
	f.Header.Set(frame.HdrId, s.id)
	sndErr := s.session.sendFrame(ctx, f)

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
