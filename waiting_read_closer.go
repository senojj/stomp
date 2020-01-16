package stomp

import (
	"io"
	"sync"
)

type waitingReadCloser struct {
	reader io.ReadCloser
	wg     *sync.WaitGroup
}

func newWaitingReadCloser(reader io.ReadCloser) *waitingReadCloser {
	var wg sync.WaitGroup
	wg.Add(1)
	return &waitingReadCloser{reader: reader, wg: &wg }
}

func (wgrc *waitingReadCloser) Read(p []byte) (int, error) {
	return wgrc.reader.Read(p)
}

func (wgrc *waitingReadCloser) wait() {
	wgrc.wg.Wait()
}

func (wgrc *waitingReadCloser) Close() error {
	closeErr := wgrc.reader.Close()
	wgrc.wg.Done()
	return closeErr
}