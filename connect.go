package stomp

import (
	"github.com/dynata/stomp/proto"
)

type Connection struct {
	version string
	session string
	server  string
}

type ConnOpt func(map[string]string)

func WithCredentials(login, passcode string) ConnOpt {
	return func(m map[string]string) {
		m[proto.HdrLogin] = login
		m[proto.HdrPasscode] = passcode
	}
}
/*
func Connect(c net.Conn, options ...ConnOpt) Connection {

}
*/
