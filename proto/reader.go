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

func readCommand(r io.Reader) (Command, error) {
	cmdReader := io.LimitReader(r, 1024)
	cmdLineReader := DelimitReader(cmdReader, byteNewLine)
	cmdLine, cmdLineRdErr := ioutil.ReadAll(cmdLineReader)

	if nil != cmdLineRdErr && io.EOF != cmdLineRdErr {
		return "", cmdLineRdErr
	}

	if len(cmdLine) == 0 {
		return "", fmt.Errorf("empty command line")
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

		if len(hdrLine) == 0 {
			break
		}
		ndx := bytes.IndexByte(hdrLine, byteColon)

		if ndx <= 0 {
			return nil, fmt.Errorf("malformed header. got %s", string(hdrLine))
		}
		name := decode(string(hdrLine[0:ndx]))
		value := decode(string(hdrLine[ndx+1:]))
		header.Append(name, value)
	}
	return header, nil
}

func (fr *FrameReader) Read() (*Frame, error) {
	command, cmdRdErr := readCommand(fr.reader)

	if nil != cmdRdErr {
		return nil, cmdRdErr
	}
	header, hdrRdErr := readHeader(fr.reader)

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
		nullTerminatedReader := DelimitReader(fr.reader, byteNull)
		body = io.MultiReader(contentLengthReader, nullTerminatedReader)
	} else {
		body = DelimitReader(fr.reader, byteNull)
	}

	return &Frame{
		Command: command,
		Header:  header,
		Body:    body,
	}, nil
}

func NewFrameReader(r io.Reader) *FrameReader {
	return &FrameReader{
		reader: bufio.NewReader(r),
	}
}
