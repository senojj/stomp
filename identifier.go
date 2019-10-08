package stomp

import (
	"strconv"
	"sync/atomic"
)

var identifier uint64

func nextId() string {
	id := atomic.AddUint64(&identifier, 1)
	return strconv.FormatUint(id, 10)
}
