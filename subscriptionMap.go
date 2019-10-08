package stomp

import "sync"

type subscriptionMap struct {
	m   map[string]func(*Message)
	mut sync.RWMutex
}

func (sm *subscriptionMap) Get(id string) (func(*Message), bool) {
	sm.mut.RLock()
	defer sm.mut.RUnlock()

	if nil == sm.m {
		return nil, false
	}
	fn, ok := sm.m[id]
	return fn, ok
}

func (sm *subscriptionMap) Set(id string, fn func(*Message)) {
	sm.mut.Lock()
	defer sm.mut.Unlock()

	if nil == sm.m {
		sm.m = make(map[string]func(*Message))
	}
	sm.m[id] = fn
}

func (sm *subscriptionMap) Del(id string) {
	sm.mut.Lock()
	defer sm.mut.Unlock()

	if nil != sm.m {
		delete(sm.m, id)
	}
}
