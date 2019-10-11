package stomp

import (
	"strconv"
	"sync/atomic"
)

var _identifier uint64

func nextId() string {
	id := atomic.AddUint64(&_identifier, 1)
	return strconv.FormatUint(id, 10)
}
