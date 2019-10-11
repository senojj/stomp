package stomp

import (
	"context"
	"github.com/dynata/stomp/proto"
)

type Transaction struct {
	id string
	session *Session
}

func (t *Transaction) Commit(ctx context.Context, options ...func(Option)) error {
	frame := proto.NewFrame(proto.CmdCommit, nil)

	for _, option := range options {
		option(Option(frame.Header))
	}
	frame.Header.Set(proto.HdrTransaction, t.id)
	return t.session.sendFrame(ctx, frame)
}

func (t *Transaction) Abort(ctx context.Context, options ...func(Option)) error {
	frame := proto.NewFrame(proto.CmdAbort, nil)

	for _, option := range options {
		option(Option(frame.Header))
	}
	frame.Header.Set(proto.HdrTransaction, t.id)
	return t.session.sendFrame(ctx, frame)
}
