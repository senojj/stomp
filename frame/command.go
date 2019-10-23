package frame

type Command string

func (c Command) String() string {
	return string(c)
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
