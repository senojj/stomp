package stomp

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestConnect(t *testing.T) {
	conn, dialErr := tls.Dial("tcp", os.Getenv("MQ_URI"), nil)

	if nil != dialErr {
		t.Fatal(dialErr)
	}
	_, connErr := Connect(
		conn,
		WithCredentials("mixr", os.Getenv("MQ_PASSWORD")),
		WithHeartBeat(0, 1000),
	)

	if nil != connErr {
		t.Fatal(connErr)
	}
}

func TestSession_Send(t *testing.T) {
	conn, dialErr := tls.Dial("tcp", os.Getenv("MQ_URI"), nil)

	if nil != dialErr {
		t.Fatal(dialErr)
	}
	session, connErr := Connect(
		conn,
		WithCredentials("mixr", os.Getenv("MQ_PASSWORD")),
		WithHeartBeat(0, 5000),
	)

	if nil != connErr {
		t.Fatal(connErr)
	}
	start := time.Now()
	for i := 0; i < 300; i++ {
		content := strings.NewReader("hello, from Stomp")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		sendErr := session.Send(
			ctx,
			"/queue/a.test",
			content,
			WithContentType("text/plain"),
			WithReceipt(strconv.Itoa(i)),
		)
		cancel()

		if nil != sendErr {
			t.Fatal(sendErr)
		}
	}
	elapsed := time.Since(start)
	fmt.Printf("time it took %ds\n", 300/(elapsed/time.Second))
}
