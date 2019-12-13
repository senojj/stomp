package stomp

import (
	"io"
	"sync"
)

type waitingReadCloser struct {
	reader io.ReadCloser
	wg     *sync.WaitGroup
}

func (wgrc *waitingReadCloser) Read(p []byte) (int, error) {
	return wgrc.reader.Read(p)
}

func (wgrc *waitingReadCloser) Close() error {
	closeErr := wgrc.reader.Close()
	wgrc.wg.Done()
	return closeErr
}