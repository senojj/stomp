## Connection Examples

### Using the Frame API
The Frame API provides you with everything you need to get started communicating with a stomp
server.
```go
package main

import (
	"errors"
	"github.com/jjware/stomp"
	"log"
	"net"
	"strings"
)

func main() {
	conn, dialErr := net.Dial("tcp", "http://somesite.com:6464")

	if dialErr != nil {
		log.Fatal(dialErr)
	}

	frmConnect := stomp.NewFrame(stomp.CmdConnect, nil)
	frmConnect.Header.Set(stomp.HdrLogin, "username")
	frmConnect.Header.Set(stomp.HdrPasscode, "password")
	frmConnect.Header.Set(stomp.HdrAcceptVersion, "2.0")

	writeErr := frmConnect.Write(conn)

	if writeErr != nil {
		log.Fatal(writeErr)
	}

	frmResponse, readErr := stomp.ReadFrame(conn)

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
The stomp connection establishment flow is synchronous, so the code hasn't become
complicated yet. However, directly writing to and reading from a connection through the Frame
API is not thread safe. Invoking `Frame::Write` concurrently on the same connection may result in
malformed frames on the wire. Invoking `stomp.ReadFrame` concurrently on the same connection may
result in errors. Even more specifically, invoking `stomp.ReadFrame` in succession on the same
connection before invoking `Close` the body of the previously read frame may result in errors. 

### Using a Handle
A Handle provides thread safe methods for writing frames to and reading frames from a connection.
A context with a timeout may be provided to a handle's methods. A Handle can be obtained by
calling the `stomp.Bind` function and passing in the desired connection.
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

	frmResponse, readErr := handle.Receive(ctx)
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