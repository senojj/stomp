package stomp

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
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

func TestSession_Send1(t *testing.T) {
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
	const sends = 50000
	start := time.Now()
	for i := 0; i < sends; i++ {
		content := strings.NewReader("hello, from Stomp")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		sendErr := session.Send(
			ctx,
			"/queue/a.test",
			content,
			WithContentType("text/plain"),
		)
		cancel()

		if nil != sendErr {
			t.Fatal(sendErr)
		}
	}
	elapsed := int64(time.Since(start) / time.Second)
	clsErr := session.Close()

	if nil != clsErr {
		fmt.Println(clsErr)
	}
	fmt.Printf("rate %ds\n", int64(float64(sends)/float64(elapsed)))
}
