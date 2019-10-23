package stomp

import (
	"context"
	"github.com/dynata/stomp/frame"
)

type Transaction struct {
	id string
	session *Session
}

func (t *Transaction) Commit(ctx context.Context, options ...func(Option)) error {
	f := frame.New(frame.CmdCommit, nil)

	for _, option := range options {
		option(Option(f.Header))
	}
	f.Header.Set(frame.HdrTransaction, t.id)
	return t.session.sendFrame(ctx, f)
}

func (t *Transaction) Abort(ctx context.Context, options ...func(Option)) error {
	f := frame.New(frame.CmdAbort, nil)

	for _, option := range options {
		option(Option(f.Header))
	}
	f.Header.Set(frame.HdrTransaction, t.id)
	return t.session.sendFrame(ctx, f)
}
