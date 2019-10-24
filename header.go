package stomp

import (
	"strings"
)

var (
	encodeReplacements = strings.NewReplacer(
		"\\", "\\\\",
		"\r", "\\r",
		"\n", "\\n",
		":", "\\c",
	)

	decodeReplacements = strings.NewReplacer(
		"\\r", "\r",
		"\\n", "\n",
		"\\c", ":",
		"\\\\", "\\",
	)
)

func encode(s string) string {
	return encodeReplacements.Replace(s)
}

func decode(s string) string {
	return decodeReplacements.Replace(s)
}

const (
	HdrContentLength = "content-length"
	HdrContentType   = "content-type"
	HdrReceipt       = "receipt"
	HdrAcceptVersion = "accept-version"
	HdrHost          = "host"
	HdrVersion       = "version"
	HdrLogin         = "login"
	HdrPasscode      = "passcode"
	HdrHeartBeat     = "heart-beat"
	HdrSession       = "session"
	HdrServer        = "server"
	HdrDestination   = "destination"
	HdrId            = "id"
	HdrAck           = "ack"
	HdrTransaction   = "transaction"
	HdrReceiptId     = "receipt-id"
	HdrSubscription  = "subscription"
	HdrMessageId     = "message-id"
	HdrMessage       = "message"
)

type Header map[string][]string

func (m Header) Append(key string, value string) {
	m[key] = append(m[key], value)
}

func (m Header) Prepend(key string, value string) {
	v := m[key]
	n := make([]string, 0, len(v)+1)
	n = append(n, value)
	n = append(n, v...)
	m[key] = n
}

func (m Header) Set(key string, value string) {
	m[key] = []string{value}
}

func (m Header) Get(key string) (string, bool) {
	if m == nil {
		return "", false
	}
	v, ok := m[key]

	if len(v) == 0 {
		return "", ok
	}
	return v[0], ok
}

func (m Header) Del(key string) {
	delete(m, key)
}
