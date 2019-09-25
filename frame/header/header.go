package header

type Header string

func (h Header) String() string {
	return string(h)
}

func Custom(value string) Header {
	return Header(value)
}

const (
	ContentLength Header = "content-length"
	ContentType   Header = "content-type"
	Receipt       Header = "receipt"
	AcceptVersion Header = "accept-version"
	Host          Header = "host"
	Version       Header = "version"
	Login         Header = "login"
	Passcode      Header = "passcode"
	HeartBeat     Header = "heart-beat"
	Session       Header = "session"
	Server        Header = "server"
	Destination   Header = "destination"
	Id            Header = "id"
	Ack           Header = "ack"
	Transaction   Header = "transaction"
	ReceiptId     Header = "receipt-id"
	Subscription  Header = "subscription"
	MessageId     Header = "message-id"
	Message       Header = "message"
)
