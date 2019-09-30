package stomp

import (
	"crypto/tls"
	"fmt"
	"os"
	"testing"
)

func TestConnect(t *testing.T) {
	conn, dialErr := tls.Dial("tcp", os.Getenv("MQ_URI"), nil)

	if nil != dialErr {
		t.Fatal(dialErr)
	}
	session, connErr := Connect(
		conn,
		WithCredentials("mixr", os.Getenv("MQ_PASSWORD")),
		WithHeartBeat(0, 1000),
	)

	if nil != connErr {
		t.Fatal(connErr)
	}
	fmt.Println(session)
}
