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

type Frame struct {
	Command Command
	Header  Header
	Body    io.ReadCloser
}

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
		Header: header,
		Body: b,
	}
}

type measurer interface {
	Len() int
}

func calculateContentLength(r io.Reader) int64 {
	if nil != r {
		if v, ok := r.(measurer); ok {
			return int64(v.Len())
		}
		return -1
	}
	return 0
}

func writeHeader(header Header, writer io.Writer) (int, error) {
	var written = 0

	for k, v := range header {
		for _, i := range v {
			b, wrtErr := fmt.Fprintf(writer, "%s:%s\n", encode(k), encode(i))

			if nil != wrtErr {
				return written, fmt.Errorf("problem writing header: %w", wrtErr)
			}
			written += b
		}
	}
	return written, nil
}

func (f *Frame) Write(w io.Writer) error {
	var bw *bufio.Writer
	_, ok := w.(io.ByteWriter)

	if !ok {
		bw = bufio.NewWriter(w)
		w = bw
	}

	_, cmdWrtErr := fmt.Fprintf(w, "%s\n", f.Command)

	if nil != cmdWrtErr {
		return cmdWrtErr
	}

	_, hdrWrtErr := writeHeader(f.Header, w)

	if nil != hdrWrtErr {
		return hdrWrtErr
	}

	_, nullWrtErr := w.Write([]byte(charNewLine))

	if nil != nullWrtErr {
		return nullWrtErr
	}

	if nil != f.Body {
		_, bdyWrtErr := io.Copy(w, f.Body)

		if nil != bdyWrtErr {
			return bdyWrtErr
		}
		closeErr := f.Body.Close()

		if nil != closeErr {
			return closeErr
		}
	}

	_, nullWrtErr = w.Write([]byte(charNull))

	if nil != nullWrtErr {
		return nullWrtErr
	}

	if bw != nil {
		return bw.Flush()
	}
	return nil
}

// ReadFrame will read an entire frame from r. The frame command line
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

type drainingCloser struct {
	io.Reader
}

func (d *drainingCloser) Close() error {
	_, err := io.Copy(ioutil.Discard, d)

	if nil != err && io.EOF != err {
		return err
	}
	return nil
}

func drainCloser(r io.Reader) io.ReadCloser {
	return &drainingCloser{r}
}

type delimitedReader struct {
	delimiter byte
	reader    io.Reader
	buf       bytes.Buffer
	done      bool
}

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

func delimitReader(r io.Reader, delimiter byte) *delimitedReader {
	return &delimitedReader{
		delimiter: delimiter,
		reader:    r,
	}
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
	cmdLineReader := delimitReader(cmdReader, byteNewLine)
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
