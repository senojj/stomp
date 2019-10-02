package proto

import (
	"fmt"
	"io"
)

type Command string

func (c Command) String() string {
	return string(c)
}

func (c Command) WriteTo(w io.Writer) (int64, error) {
	b, e := fmt.Fprintf(w, "%s\n", c)
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
