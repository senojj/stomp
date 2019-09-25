package command

import (
	"fmt"
	"io"
)

type Command string

func (c Command) String() string {
	return string(c)
}

func (c Command) WriteTo(w io.Writer) (int64, error) {
	out := fmt.Sprintf("%s\n", c)
	b, e := w.Write([]byte(out))
	return int64(b), e
}

const (
	Connect     Command = "CONNECT"
	Stomp       Command = "STOMP"
	Connected   Command = "CONNECTED"
	Send        Command = "SEND"
	Subscribe   Command = "SUBSCRIBE"
	Unsubscribe Command = "UNSUBSCRIBE"
	Ack         Command = "ACK"
	Nack        Command = "NACK"
	Begin       Command = "BEGIN"
	Commit      Command = "COMMIT"
	Abort       Command = "ABORT"
	Disconnect  Command = "DISCONNECT"
	Message     Command = "MESSAGE"
	Receipt     Command = "RECEIPT"
	Error       Command = "ERROR"
)
