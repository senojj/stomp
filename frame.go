package stomp

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
)

const (
	defaultMaxHeaderBytes = 1 << 20 // 1 MB
)

// A Frame represents a STOMP frame received or sent by
// a server or client.
type Frame struct {
	// Command specifies the STOMP command.
	Command Command

	// Header contains the request header fields.
	Header Header

	// Body is the frame's body. A nil body means the frame
	// has no body. A frame resulting from a ReadFrame must
	// have its body closed explicitly. Also, a frame resulting
	// from a ReadFrame will always have a non-nil body, but
	// will return an io.EOF immediately when read.
	Body io.ReadCloser
}

// NewFrame returns a new Frame given a command, and an optional
// body. If the provided body is also an io.Closer, the returned
// Frame.Body is set to body, otherwise body will be wrapped in
// an ioutil.NopCloser. If body is also a measurer, the content-
// type header will be pre-populated for the frame.
func NewFrame(command Command, body io.Reader) *Frame {
	header := make(Header)
	length := calculateContentLength(body)

	if length > 0 {
		header.Set(HdrContentLength, strconv.FormatInt(length, 10))
	}

	var b io.ReadCloser

	if nil != body {
		readCloser, ok := body.(io.ReadCloser)

		if ok {
			b = readCloser
		} else {
			b = ioutil.NopCloser(body)
		}
	}

	return &Frame{
		Command: command,
		Header:  header,
		Body:    b,
	}
}

// The measurer interface wraps the basic Len method.
// Len returns the number of bytes of the unread portion of the
// underlying type.
type measurer interface {
	Len() int
}

// calculateContentLength returns the length, in bytes, of r
// if the value of r satisfies the measurer interface. If r
// is not measurable, -1 is returned. If r is nil, 0 is returned.
func calculateContentLength(r io.Reader) int64 {
	if nil != r {
		if v, ok := r.(measurer); ok {
			return int64(v.Len())
		}
		return -1
	}
	return 0
}

// WriteTo writes a STOMP frame, which is the command, header, and body,
// in wire format. If Body is present, WriteTo closes Body once it
// has been written in full.
func (f *Frame) WriteTo(w io.Writer) (int64, error) {
	var bw *bufio.Writer
	var totalBytesWrt int64

	if _, ok := w.(io.ByteWriter); !ok {
		bw = bufio.NewWriter(w)
		w = bw
	}

	cmdbyt, cmdWrtErr := fmt.Fprintf(w, "%s\n", f.Command)

	if nil != cmdWrtErr {
		return totalBytesWrt, cmdWrtErr
	}
	totalBytesWrt += int64(cmdbyt)

	hdrbyt, hdrWrtErr := f.Header.WriteTo(w)

	if nil != hdrWrtErr {
		return totalBytesWrt, hdrWrtErr
	}
	totalBytesWrt += hdrbyt

	nlbyt, nullWrtErr := w.Write([]byte(charNewLine))

	if nil != nullWrtErr {
		return totalBytesWrt, nullWrtErr
	}
	totalBytesWrt += int64(nlbyt)

	if nil != f.Body {
		bodybyt, bdyWrtErr := io.Copy(w, f.Body)

		if nil != bdyWrtErr {
			return totalBytesWrt, bdyWrtErr
		}
		totalBytesWrt += bodybyt
		closeErr := f.Body.Close()

		if nil != closeErr {
			return totalBytesWrt, closeErr
		}
	}

	nlbyt, nullWrtErr = w.Write([]byte(charNull))

	if nil != nullWrtErr {
		return totalBytesWrt, nullWrtErr
	}
	totalBytesWrt += int64(nlbyt)

	if bw != nil {
		return totalBytesWrt, bw.Flush()
	}
	return totalBytesWrt, nil
}

// ReadFrame will read an entire frame from r. The frame command
// and header lines are restricted to specified sizes to guard
// against malicious frame writes. The frame body is left unread on
// the stream, and can be read up until either the content length
// (when specified) is reached, or a null character is encountered.
// The body of the frame must be explicitly closed by the reader in
// order to remove the frame's contents from the stream prior to the
// next read. A return value of (nil, nil) indicates that a heart-beat
// was likely received.
func ReadFrame(r io.Reader) (*Frame, error) {
	nullTerminatedReader := delimitReader(r, byteNull)
	command, cmdRdErr := readCommand(nullTerminatedReader)

	if nil != cmdRdErr {
		return nil, cmdRdErr
	}

	if "" == command {
		return nil, nil
	}
	header, hdrRdErr := readHeader(nullTerminatedReader)

	if nil != hdrRdErr {
		return nil, hdrRdErr
	}
	contentLengths, hasContentLength := header[HdrContentLength]
	var body io.Reader

	if hasContentLength {
		contentLength, convErr := strconv.ParseInt(contentLengths[0], 10, 64)

		if nil != convErr {
			return nil, convErr
		}
		contentLengthReader := io.LimitReader(r, contentLength)
		body = io.MultiReader(contentLengthReader, nullTerminatedReader)
	} else {
		body = nullTerminatedReader
	}

	return &Frame{
		Command: command,
		Header:  header,
		Body:    drainCloser(body),
	}, nil
}

// drainingCloser satisfies the io.ReadCloser interface.
type drainingCloser struct {
	io.Reader
}

// Close will drain the remaining bytes to be read from
// the internal reader to devNull.
func (d *drainingCloser) Close() error {
	_, err := io.Copy(ioutil.Discard, d)

	if nil != err && io.EOF != err {
		return err
	}
	return nil
}

// drainCloser creates a new drainingCloser, wrapping
// the given reader r.
func drainCloser(r io.Reader) io.ReadCloser {
	return &drainingCloser{r}
}

// delimitedReader satisfies the io.Reader interface.
type delimitedReader struct {
	delimiter byte
	reader    io.Reader
	buf       bytes.Buffer
	done      bool
}

// Read reads up to len(p) bytes into p from the internal
// buffer until the delimiter or io.EOF has been reached.
// If the delimiter is reached, successive calls to Read
// will result in an io.EOF. The delimiter is not written
// into p.
func (d *delimitedReader) Read(p []byte) (int, error) {
	if d.done {
		return 0, io.EOF
	}
	totalRead := 0
	buf := make([]byte, 1)

	for i := 0; i < len(p); i++ {
		read, rdErr := d.reader.Read(buf)

		if read < 0 {
			return 0, fmt.Errorf("negative read")
		}

		if read > 0 {
			if buf[0] == d.delimiter {
				d.done = true
				return totalRead, io.EOF
			}
			totalRead += read
			p[i] = buf[0]
		}

		if nil != rdErr {
			return totalRead, rdErr
		}
	}
	return totalRead, nil
}

// delimitReader creates a new delimitedReader instance given an
// io.Reader and a byte delimiter.
func delimitReader(r io.Reader, delimiter byte) *delimitedReader {
	return &delimitedReader{
		delimiter: delimiter,
		reader:    r,
	}
}

// stripCarriageReturn removes an existing
// carriage return from the end of a byte slice.
func stripCarriageReturn(b []byte) []byte {
	ndx := bytes.LastIndexByte(b, byteCarriageReturn)

	if ndx < 0 {
		return b
	}
	return b[:ndx]
}

// readCommand reads a frame's command line from the
// provided io.Reader. The command line will be read
// until a new line character is encountered, 1024 bytes
// have been read, or an io.EOF is encountered. If a
// carriage return exists at the end of the command line,
// it will be stripped.
func readCommand(r io.Reader) (Command, error) {
	cmdReader := io.LimitReader(r, 1024)
	cmdLineReader := delimitReader(cmdReader, byteNewLine)
	cmdLine, cmdLineRdErr := ioutil.ReadAll(cmdLineReader)
	cmdLine = stripCarriageReturn(cmdLine)
	return Command(cmdLine), cmdLineRdErr
}

// readHeader reads a frame's header from the provided io.Reader.
// Each header line will be read until a new line character is
// encountered, defaultMaxHeaderBytes has been read, or an io.EOF
// is encountered. Each header line will have an existing carriage
// return stripped. Each header line will have its name and value
// decoded according to the STOMP specification.
func readHeader(r io.Reader) (Header, error) {
	header := make(Header)
	hdrReader := io.LimitReader(r, defaultMaxHeaderBytes)

	for {
		hdrLineReader := delimitReader(hdrReader, byteNewLine)
		hdrLine, hdrLineErr := ioutil.ReadAll(hdrLineReader)

		if nil != hdrLineErr && io.EOF != hdrLineErr {
			return nil, hdrLineErr
		}
		hdrLine = stripCarriageReturn(hdrLine)

		if len(hdrLine) == 0 {
			break
		}
		ndx := bytes.IndexByte(hdrLine, byteColon)

		if ndx <= 0 {
			return nil, fmt.Errorf("malformed header. got %v", hdrLine)
		}
		name := decode(string(hdrLine[0:ndx]))
		value := decode(string(hdrLine[ndx+1:]))
		header.Append(name, value)
	}
	return header, nil
}
