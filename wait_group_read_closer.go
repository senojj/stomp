package stomp

import (
	"io"
	"sync"
)

type waitGroupReadCloser struct {
	reader io.ReadCloser
	wg     *sync.WaitGroup
}

func (wgrc *waitGroupReadCloser) Read(p []byte) (int, error) {
	return wgrc.reader.Read(p)
}

func (wgrc *waitGroupReadCloser) Close() error {
	closeErr := wgrc.reader.Close()
	wgrc.wg.Done()
	return closeErr
}
