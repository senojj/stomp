package stomp

import "errors"

var (
	ErrSessionClosed = errors.New("session is closed")
)
