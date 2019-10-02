package stomp

import "sync"

type receiptMap struct {
	m   map[string]chan<- error
	mut sync.RWMutex
}

func (rm *receiptMap) Get(id string) (chan<- error, bool) {
	rm.mut.RLock()
	defer rm.mut.RUnlock()

	if nil == rm.m {
		return nil, false
	}
	ch, ok := rm.m[id]
	return ch, ok
}

func (rm *receiptMap) Set(id string, ch chan<- error) {
	rm.mut.Lock()
	defer rm.mut.Unlock()

	if nil == rm.m {
		rm.m = make(map[string]chan<- error)
	}
	rm.m[id] = ch
}

func (rm *receiptMap) Del(id string) {
	rm.mut.Lock()
	defer rm.mut.Unlock()

	if nil != rm.m {
		delete(rm.m, id)
	}
}
