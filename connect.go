package stomp

import (
	"context"
	"errors"
	"strings"
)

func Connect(t *Transport, username, password string) error {
	return ConnectWithContext(context.Background(), t, username, password)
}

func ConnectWithContext(ctx context.Context, t *Transport, username, password string) error {
	f := NewFrame(CmdConnect, nil)
	f.Header.Set(HdrLogin, username)
	f.Header.Set(HdrPasscode, password)

	sendErr := t.Send(ctx, f)

	if nil != sendErr {
		return sendErr
	}
	resp, readErr := t.Read(ctx)

	if nil != readErr {
		return readErr
	}

	if resp.Command != CmdConnected {
		message, ok := resp.Header["message"]

		if ok {
			return errors.New(strings.Join(message, ":"))
		}
		return errors.New("unknown error")
	}
	closeErr := resp.Body.Close()

	if nil != closeErr {
		return closeErr
	}
	return nil
}
