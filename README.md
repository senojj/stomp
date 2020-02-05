## Connection Example
```go
package main

import (
	"context"
	"errors"
	"github.com/jjware/stomp"
	"log"
	"net"
	"strings"
	"time"
)

func main() {
	conn, dialErr := net.Dial("tcp", "http://someuri.com:61614")

	if dialErr != nil {
		log.Fatal(dialErr)
	}
	handle := stomp.Bind(conn)

	frmConnect := stomp.NewFrame(stomp.CmdConnect, nil)
	frmConnect.Header.Set(stomp.HdrLogin, "username")
	frmConnect.Header.Set(stomp.HdrPasscode, "password")
	frmConnect.Header.Set(stomp.HdrAcceptVersion, "2.0")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	sendErr := handle.Send(ctx, frmConnect)

	if sendErr != nil {
		log.Fatal(sendErr)
	}

	frmResponse, readErr := handle.Read(ctx)
	cancel()

	if readErr != nil {
		log.Fatal(readErr)
	}

	if frmResponse.Command == stomp.CmdError {

		if message, ok := frmResponse.Header[stomp.HdrMessage]; ok {
			log.Fatal(errors.New(strings.Join(message, ":")))
		}
		log.Fatal("unknown error")
	}

	if frmResponse.Command != stomp.CmdConnected {
		log.Fatalf("unexpected frame command: %s", frmResponse.Command)
	}

	// connection established
}
```