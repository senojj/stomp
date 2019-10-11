package stomp

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
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
			WithReceipt(),
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

	subscription, subErr := session.Subscribe(ctx, "/queue/a.test")
	cancel()

	if nil != subErr {
		t.Fatal(subErr)
	}

	go func() {
		for message := range subscription.C {
			body, rdErr := ioutil.ReadAll(message.Body)
			closeErr := message.Body.Close()

			if nil != rdErr {
				t.Fatal(rdErr)
			}

			if nil != closeErr {
				t.Fatal(closeErr)
			}
			t.Logf("received message - %s\n", string(body))
		}
	}()

	t.Logf("successfully subscribed...")
	time.Sleep(5*time.Second)
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	unsubErr := subscription.Unsubscribe(ctx, WithReceipt())
	cancel()

	if nil != unsubErr {
		t.Fatal(unsubErr)
	}
	clsErr := session.Close()

	if nil != clsErr {
		t.Fatal(clsErr)
	}
}

func TestSession_SendSubscribe(t *testing.T) {
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

	subscription, subErr := session.Subscribe(ctx, "/queue/a.test")
	cancel()

	if nil != subErr {
		t.Fatal(subErr)
	}

	go func() {
		for message := range subscription.C {
			body, rdErr := ioutil.ReadAll(message.Body)
			closeErr := message.Body.Close()

			if nil != rdErr {
				t.Fatal(rdErr)
			}

			if nil != closeErr {
				t.Fatal(closeErr)
			}
			t.Logf("received message - %s\n", string(body))
		}
	}()

	start := time.Now().Add(5*time.Second)

	for time.Now().Before(start) {
		content := bytes.NewReader([]byte("hello, from stomp"))
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		sndErr := session.Send(ctx, "/queue/a.test", content)
		cancel()

		if nil != sndErr {
			t.Fatal(sndErr)
		}
	}
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	unsubErr := subscription.Unsubscribe(ctx, WithReceipt())
	cancel()

	if nil != unsubErr {
		t.Fatal(unsubErr)
	}
	clsErr := session.Close()

	if nil != clsErr {
		t.Fatal(clsErr)
	}
}

func TestSession_Begin(t *testing.T) {
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
	transaction, trnErr := session.Begin(ctx, WithReceipt())
	cancel()

	if nil != trnErr {
		t.Fatal(trnErr)
	}

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	subscription, subErr := session.Subscribe(ctx, "/queue/a.test", WithAck(AckClientIndividual), WithReceipt())
	cancel()

	if nil != subErr {
		t.Fatal(subErr)
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		for message := range subscription.C {
			closeErr := message.Body.Close()

			if nil != closeErr {
				t.Fatal(closeErr)
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				ackErr := session.Ack(ctx, message, WithTransaction(transaction))
				cancel()

				if nil != ackErr {
					t.Error(ackErr)
				}
			}()
		}
		wg.Done()
	}()

	time.Sleep(10*time.Second)

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	unsubErr := subscription.Unsubscribe(ctx, WithReceipt())
	cancel()

	if nil != unsubErr {
		t.Fatal(unsubErr)
	}

	wg.Wait()

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	abrtErr := transaction.Commit(ctx, WithReceipt())
	cancel()

	if nil != abrtErr {
		t.Fatal(abrtErr)
	}
	clsErr := session.Close()

	if nil != clsErr {
		t.Fatal(clsErr)
	}
}
