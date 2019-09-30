package proto

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
)

const (
	DefaultMaxHeaderBytes = 1 << 20 // 1 MB
)

type MeasurableReader interface {
	io.Reader
	Len() int
}

type flushCloser struct {
	io.Reader
}

func (fc *flushCloser) Close() error {
	_, err := ioutil.ReadAll(fc)

	if nil != err && io.EOF != err {
		return err
	}
	return nil
}

func FlushCloser(r io.Reader) io.ReadCloser {
	return &flushCloser{r}
}

type DelimitedReader struct {
	delimiter byte
	reader    io.Reader
	buf       bytes.Buffer
	done      bool
}

func (d *DelimitedReader) Read(p []byte) (int, error) {
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
		totalRead += read

		if read > 0 {
			p[i] = buf[0]

			if buf[0] == d.delimiter {
				d.done = true
				return i, io.EOF
			}
		}

		if nil != rdErr {
			return i, rdErr
		}
	}
	return totalRead, nil
}

func DelimitReader(r io.Reader, delimiter byte) *DelimitedReader {
	return &DelimitedReader{
		delimiter: delimiter,
		reader:    r,
	}
}

type FrameReader struct {
	reader *bufio.Reader
}

func stripCarriageReturn(b []byte) []byte {
	ndx := bytes.LastIndexByte(b, byteCarriageReturn)

	if ndx < 0 {
		return b
	}
	return b[:ndx]
}

func readCommand(r io.Reader) (Command, error) {
		cmdReader := io.LimitReader(r, 1024)
		cmdLineReader := DelimitReader(cmdReader, byteNewLine)
		cmdLine, cmdLineRdErr := ioutil.ReadAll(cmdLineReader)

		cmdLine = stripCarriageReturn(cmdLine)

		if nil != cmdLineRdErr && io.EOF != cmdLineRdErr {
			return "", cmdLineRdErr
		}

		if len(cmdLine) == 0 {
			return "", nil
		}
		return Command(cmdLine), nil
}

func readHeader(r io.Reader) (Header, error) {
	header := make(Header)
	hdrReader := io.LimitReader(r, DefaultMaxHeaderBytes)

	for {
		hdrLineReader := DelimitReader(hdrReader, byteNewLine)
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

// Read will read an entire frame from the internal reader
// as a ServerFrame. The frame command line and header lines
// are restricted to specified sizes to guard against malicious
// frame writes. The frame body is left unread on the stream,
// and can be read up until either the content length (when specified)
// is reached, or a null character is encountered. The body of
// a ServerFrame must be explicitly closed by the reader in order
// to remove the frame's contents from the stream prior to the next
// read. A return value of (nil, nil) indicates that a heart-beat
// was received.
func (fr *FrameReader) Read() (*ServerFrame, error) {
	nullTerminatedReader := DelimitReader(fr.reader, byteNull)
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
		contentLengthReader := io.LimitReader(fr.reader, contentLength)
		body = io.MultiReader(contentLengthReader, nullTerminatedReader)
	} else {
		body = nullTerminatedReader
	}

	return &ServerFrame{
		Command: command,
		Header:  header,
		Body:    FlushCloser(body),
	}, nil
}

func NewFrameReader(r io.Reader) *FrameReader {
	return &FrameReader{
		reader: bufio.NewReader(r),
	}
}
