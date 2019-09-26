package stomp

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
	CmdConnect     Command = "CONNECT"
	CmdStomp       Command = "STOMP"
	CmdConnected   Command = "CONNECTED"
	CmdSend        Command = "SEND"
	CmdSubscribe   Command = "SUBSCRIBE"
	CmdUnsubscribe Command = "UNSUBSCRIBE"
	CmdAck         Command = "ACK"
	CmdNack        Command = "NACK"
	CmdBegin       Command = "BEGIN"
	CmdCommit      Command = "COMMIT"
	CmdAbort       Command = "ABORT"
	CmdDisconnect  Command = "DISCONNECT"
	CmdMessage     Command = "MESSAGE"
	CmdReceipt     Command = "RECEIPT"
	CmdError       Command = "ERROR"
)
