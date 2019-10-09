package stomp

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
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
	session, connErr := Connect(
		conn,
		WithCredentials("mixr", os.Getenv("MQ_PASSWORD")),
		WithHeartBeat(0, 1000),
	)

	if nil != connErr {
		t.Fatal(connErr)
	}
	clsErr := session.Close()

	if nil != clsErr {
		t.Fatal(clsErr)
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

func TestSession_Send2(t *testing.T) {
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
	const sends = 500
	start := time.Now()
	for i := 0; i < sends; i++ {
		content := strings.NewReader("hello, from Stomp")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		sendErr := session.Send(
			ctx,
			"/queue/a.test",
			content,
			WithContentType("text/plain"),
			WithReceipt("send-" + strconv.Itoa(i)),
		)
		cancel()

		if nil != sendErr {
			t.Error(sendErr)
		}
	}
	elapsed := int64(time.Since(start) / time.Second)
	clsErr := session.Close()

	if nil != clsErr {
		fmt.Println(clsErr)
	}
	fmt.Printf("rate %ds\n", int64(float64(sends)/float64(elapsed)))
}

func TestSession_Subscribe(t *testing.T) {
	conn, dialErr := tls.Dial("tcp", os.Getenv("MQ_URI"), nil)

	if nil != dialErr {
		t.Fatal(dialErr)
	}
	session, connErr := Connect(
		conn,
		WithCredentials("mixr", os.Getenv("MQ_PASSWORD")),
	)

	if nil != connErr {
		t.Fatal(connErr)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	subscription, subErr := session.Subscribe(ctx, "/queue/a.test", func(message Message) {
		body, rdErr := ioutil.ReadAll(message.Body)
		closeErr := message.Body.Close()

		if nil != rdErr {
			t.Fatal(rdErr)
		}

		if nil != closeErr {
			t.Fatal(closeErr)
		}
		fmt.Printf("received message - %s\n", string(body))
	})
	cancel()

	if nil != subErr {
		t.Fatal(subErr)
	}
	time.Sleep(5*time.Second)
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	unsubErr := subscription.Unsubscribe(ctx)
	cancel()

	if nil != unsubErr {
		t.Fatal(unsubErr)
	}
	clsErr := session.Close()

	if nil != clsErr {
		fmt.Println(clsErr)
	}
}
