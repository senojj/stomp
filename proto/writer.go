package proto

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

type FrameWriter struct {
	writer *bufio.Writer
}

func NewFrameWriter(w io.Writer) *FrameWriter {
	return &FrameWriter{
		writer: bufio.NewWriterSize(w, 4096),
	}
}

func calculateContentLength(r io.Reader) int {
	if nil != r {
		if v, ok := r.(MeasurableReader); ok {
			return v.Len()
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

func (fw *FrameWriter) Write(frame *ClientFrame) (int64, error) {
	var written int64 = 0

	cmdBytes, cmdWrtErr := fmt.Fprintf(fw.writer, "%s\n", frame.Command)

	if nil != cmdWrtErr {
		return written, cmdWrtErr
	}
	written += int64(cmdBytes)

	contentLength := calculateContentLength(frame.Body)

	if contentLength > -1 {
		frame.Header.Set(HdrContentLength, strconv.Itoa(contentLength))
	}
	hdrBytes, hdrWrtErr := writeHeader(frame.Header, fw.writer)

	if nil != hdrWrtErr {
		return written, hdrWrtErr
	}
	written += int64(hdrBytes)

	nullBytes, nullWrtErr := fw.writer.Write([]byte(charNewLine))

	if nil != nullWrtErr {
		return written, nullWrtErr
	}
	written += int64(nullBytes)

	if nil != frame.Body {
		bdyb, bdyWrtErr := io.Copy(fw.writer, frame.Body)

		if nil != bdyWrtErr {
			return written, bdyWrtErr
		}
		written += bdyb
	}

	nullBytes, nullWrtErr = fw.writer.Write([]byte(charNull))

	if nil != nullWrtErr {
		return written, nullWrtErr
	}
	written += int64(nullBytes)
	flErr := fw.writer.Flush()

	if nil != flErr {
		return written, flErr
	}
	return written, nil
}
