package stomp

import (
	"strconv"
	"sync/atomic"
)

type Sequence struct {
	value uint64
}

func (s *Sequence) Next() string {
	v := atomic.AddUint64(&s.value, 1)
	return strconv.FormatUint(v, 10)
}